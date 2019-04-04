/*
 * Copyright (c) 2018 ICON Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package foundation.icon.test.score;

import foundation.icon.icx.IconService;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Address;
import foundation.icon.icx.data.TransactionResult;
import foundation.icon.icx.transport.jsonrpc.RpcObject;
import foundation.icon.icx.transport.jsonrpc.RpcValue;
import foundation.icon.test.common.Constants;
import foundation.icon.test.common.Utils;

import java.io.IOException;
import java.math.BigInteger;

import static foundation.icon.test.common.Env.LOG;

public class CrowdSaleScore extends Score {
    private static final String PATH = Constants.SCORE_ROOT + "crowdSale.zip";

    public static CrowdSaleScore mustDeploy(IconService service, Wallet wallet, BigInteger nid,
                                            BigInteger goalInIcx, Address tokenScore, int durationInBlocks)
            throws IOException, TransactionFailureException
    {
        RpcObject params = new RpcObject.Builder()
                .put("_fundingGoalInIcx", new RpcValue(goalInIcx))
                .put("_tokenScore", new RpcValue(tokenScore))
                .put("_durationInBlocks", new RpcValue(BigInteger.valueOf(durationInBlocks)))
                .build();
        return new CrowdSaleScore(
                service,
                Score.mustDeploy(service, wallet, PATH, params),
                nid
        );
    }

    public CrowdSaleScore(IconService iconService, Address scoreAddress, BigInteger nid) {
        super(iconService, scoreAddress, nid);
    }

    public TransactionResult checkGoalReached(Wallet wallet) throws IOException {
        return invokeAndWaitResult(wallet,
                "checkGoalReached", null, null, STEPS_DEFAULT);
    }

    public void ensureCheckGoalReached(Wallet wallet) throws Exception {
        while (true) {
            TransactionResult result = checkGoalReached(wallet);
            if (!Constants.STATUS_SUCCESS.equals(result.getStatus())) {
                throw new IOException("Failed to execute checkGoalReached.");
            }
            TransactionResult.EventLog event = Utils.findEventLogWithFuncSig(result, scoreAddress, "GoalReached(Address,int)");
            if (event != null) {
                break;
            }
            LOG.info("Sleep 1 second.");
            Thread.sleep(1000);
        }
    }

    public TransactionResult safeWithdrawal(Wallet wallet) throws IOException {
        return invokeAndWaitResult(wallet, "safeWithdrawal", null, null, STEPS_DEFAULT);
    }
}