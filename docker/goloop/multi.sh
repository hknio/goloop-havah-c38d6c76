#!/bin/sh
set -e

PRE_PWD=$(pwd)
WORKDIR=$(dirname "$(readlink -f ${0})")
cd $WORKDIR

REPO_GOLOOP=${REPO_GOLOOP:-goloop}
GOLOOP_DATA=${GOLOOP_DATA:-/goloop/data}
GOLOOP_DOCKER_REPLICAS=${GOLOOP_DOCKER_REPLICAS:-4}
GOLOOP_DOCKER_NETWORK=${GOLOOP_DOCKER_NETWORK:-goloop_net}
GOLOOP_DOCKER_VOLUME=${GOLOOP_DOCKER_VOLUME:-goloop_data}
GOLOOP_DOCKER_MOUNT=${GOLOOP_DOCKER_MOUNT:-${GOLOOP_DATA}}
GOLOOP_DOCKER_PREFIX=${GOLOOP_DOCKER_PREFIX:-goloop}

GSTOOL=${GSTOOL:-../../bin/gstool}

function create(){
    docker network create --driver overlay --attachable ${GOLOOP_DOCKER_NETWORK} || echo "already created ${GOLOOP_DOCKER_NETWORK}"
    docker volume create ${GOLOOP_DOCKER_VOLUME} || echo "already created ${GOLOOP_DOCKER_VOLUME}"
    
    for i in $(seq 0 $((${GOLOOP_DOCKER_REPLICAS}-1)));do 
        GOLOOP_NODE_DIR="${GOLOOP_DATA}/${i}"
        GOLOOP_CONFIG="${GOLOOP_NODE_DIR}/config.json"
        GOLOOP_KEY_STORE="${GOLOOP_NODE_DIR}/keystore.json"
        GOLOOP_KEY_SECRET="${GOLOOP_NODE_DIR}/secret"
        GOLOOP_LOGFILE="${GOLOOP_NODE_DIR}/goloop.log"
        
        # keystore
        mkdir -p $(dirname ${GOLOOP_KEY_SECRET})
        echo -n "${GOLOOP_DOCKER_PREFIX}-${i}" > ${GOLOOP_KEY_SECRET}
        echo "${GSTOOL} ks gen -o ${GOLOOP_KEY_STORE} -p \$(cat ${GOLOOP_KEY_SECRET})"
        ${GSTOOL} ks gen -o "${GOLOOP_KEY_STORE}" -p $(cat ${GOLOOP_KEY_SECRET})
        
        docker run -d \
          --mount type=volume,src=${GOLOOP_DOCKER_VOLUME},dst=${GOLOOP_DOCKER_MOUNT} \
          --network ${GOLOOP_DOCKER_NETWORK} \
          --network-alias ${GOLOOP_DOCKER_PREFIX}-${i} \
          --name ${GOLOOP_DOCKER_PREFIX}-${i} \
          --hostname ${GOLOOP_DOCKER_PREFIX}-${i} \
          --env TASK_SLOT=${i} \
          --env GOLOOP_NODE_DIR=${GOLOOP_NODE_DIR} \
          --env GOLOOP_CONFIG=${GOLOOP_CONFIG} \
          --env GOLOOP_KEY_STORE=${GOLOOP_KEY_STORE} \
          --env GOLOOP_KEY_SECRET=${GOLOOP_KEY_SECRET} \
          --env GOLOOP_LOGFILE=${GOLOOP_LOGFILE} \
          --env GOLOOP_P2P=${GOLOOP_DOCKER_PREFIX}-${i}:8080 \
          ${REPO_GOLOOP}
    done
}

function join(){
    local GENESIS_TEMPLATE=${1:-${GOLOOP_DATA}/genesis/genesis.json}
    local GOD_KEYSTORE=${2}
    
    # collect node addresses
    for i in $(seq 0 $((${GOLOOP_DOCKER_REPLICAS}-1)));do 
        ADDRESS=$(docker exec ${GOLOOP_DOCKER_PREFIX}-${i} goloop system --format "{{.Address}}")
        VALIDATORS="${VALIDATORS} -v ${ADDRESS}"
        ADDRESSES="${ADDRESSES} ${ADDRESS}"
    done
    
    # god keystore
    if [ "${GOD_KEYSTORE}" != "" ] && [ ! -f ${GOD_KEYSTORE} ]; then
      mkdir -p $(dirname ${GOD_KEYSTORE})
      ${GSTOOL} ks gen -o ${GOD_KEYSTORE}
    fi
    GSTOOL_CMD="${GSTOOL} gn --god ${GOD_KEYSTORE}"
    # genesis
    if [ ! -f ${GENESIS_TEMPLATE} ];then
        mkdir -p $(dirname ${GENESIS_TEMPLATE})
        GSTOOL_CMD="${GSTOOL_CMD} -o ${GENESIS_TEMPLATE} gen ${ADDRESSES}"
    else
        GSTOOL_CMD="${GSTOOL_CMD} ${VALIDATORS} edit ${GENESIS_TEMPLATE}"
    fi
    echo ${GSTOOL_CMD}
    ${GSTOOL_CMD}
    
    for i in $(seq 0 $((${GOLOOP_DOCKER_REPLICAS}-1)));do 
        docker exec ${GOLOOP_DOCKER_PREFIX}-${i} goloop chain join --genesis_template ${GENESIS_TEMPLATE} --seed "${GOLOOP_DOCKER_PREFIX}-0":8080 1
    done
}

function start(){
    for i in $(seq 0 $((${GOLOOP_DOCKER_REPLICAS}-1)));do 
        docker exec ${GOLOOP_DOCKER_PREFIX}-${i} goloop chain start 1
    done
}

function env(){
    local ENVFILE=${1}
    cp ${ENVFILE} ${ENVFILE}.backup
    grep "^chain" ${ENVFILE}.backup > ${ENVFILE}
    for i in $(seq 0 $((${GOLOOP_DOCKER_REPLICAS}-1)));do
        echo -e "\nnode${i}.url=http://${GOLOOP_DOCKER_PREFIX}-${i}:9080\nnode${i}.channel0.nid=1\nnode${i}.channel0.name=1" >> ${ENVFILE}
    done
}

function rm(){
    for i in $(seq 0 $((${GOLOOP_DOCKER_REPLICAS}-1)));do
        echo "docker stop $(docker stop ${GOLOOP_DOCKER_PREFIX}-${i})"
        echo "docker rm $(docker rm ${GOLOOP_DOCKER_PREFIX}-${i})"
    done
    echo "docker network rm $(docker network rm ${GOLOOP_DOCKER_NETWORK})"
    echo "docker volume rm $(docker volume rm ${GOLOOP_DOCKER_VOLUME})"
}

case $1 in
create)
  create
;;
join)
  join $2 $3
;;
start)
  start
;;
env)
  env $2
;;
rm)
  rm
;;
*)
  echo "Usage: $0 [create,join,start,env,rm]"
  echo "  create: $0 create"
  echo "  join: $0 join [GENESIS_TEMPLATE] [GOD_KEYSTORE]"
  echo "  start: $0 start"
  echo "  env: $0 env [ENV_PROPERTIES]"
  echo "  rm: $0 rm"
  cd $PRE_PW
  exit 1
;;
esac

cd $PRE_PW