/*
 * Copyright 2019 ICON Foundation
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

package foundation.icon.ee.score;

import foundation.icon.ee.ipc.Proxy;
import foundation.icon.ee.types.Bytes;
import org.aion.avm.core.IExternalState;
import org.aion.avm.core.util.Helpers;
import org.aion.types.AionAddress;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.File;
import java.io.IOException;
import java.math.BigInteger;
import java.nio.file.Files;
import java.nio.file.NoSuchFileException;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Arrays;

public class ExternalState implements IExternalState {
    private static final Logger logger = LoggerFactory.getLogger(ExternalState.class);
    private static final String TRANSFORMED_JAR = "transformed.jar";
    private static final String OBJECT_GRAPH = "graph";

    private final Proxy proxy;
    private final Path jarPath;
    private final long blockNumber;
    private final long blockTimestamp;
    private final Path parentPath;

    ExternalState(Proxy proxy, String code, BigInteger blockNumber, BigInteger blockTimestamp) {
        this.proxy = proxy;
        this.jarPath = Paths.get(code);
        this.blockNumber = blockNumber.longValue();
        this.blockTimestamp = blockTimestamp.longValue() / 1000; // micro to milli conversion
        this.parentPath = jarPath.getParent();
    }

    private void writeFile(String filename, byte[] data) {
        Path outFile = new File(parentPath.toFile(), filename).toPath();
        try {
            Files.write(outFile, data);
        } catch (IOException e) {
            throw new RuntimeException(e.getMessage());
        }
    }

    private byte[] readFile(String filename) {
        Path inFile = new File(parentPath.toFile(), filename).toPath();
        try {
            return Files.readAllBytes(inFile);
        } catch (NoSuchFileException e) {
            throw new RuntimeException("No such file: " + e.getMessage());
        } catch (IOException e) {
            throw new RuntimeException(e.getMessage());
        }
    }

    @Override
    public void commit() {
        logger.debug("[commit]");
        throw new RuntimeException("not implemented");
    }

    @Override
    public void commitTo(IExternalState externalState) {
        logger.debug("[commitTo] {}", externalState);
        throw new RuntimeException("not implemented");
    }

    @Override
    public IExternalState newChildExternalState() {
        logger.debug("[newChildExternalState]");
        throw new RuntimeException("not implemented");
    }

    @Override
    public void createAccount(AionAddress address) {
        logger.debug("[createAccount] {}", address);
        throw new RuntimeException("not implemented");
    }

    @Override
    public boolean hasAccountState(AionAddress address) {
        logger.debug("[hasAccountState] {}", address);
        throw new RuntimeException("not implemented");
    }

    @Override
    public byte[] getCode(AionAddress address) {
        logger.debug("[getCode] {}", address);
        throw new RuntimeException("not implemented");
    }

    @Override
    public void putCode(AionAddress address, byte[] code) {
        logger.debug("[putCode] {} len={}", address, code.length);
        // just ignore this
    }

    @Override
    public byte[] getTransformedCode(AionAddress address) {
        logger.debug("[getTransformedCode] {}", address);
        return readFile(TRANSFORMED_JAR);
    }

    @Override
    public void setTransformedCode(AionAddress address, byte[] code) {
        logger.debug("[setTransformedCode] {} len={}", address, code.length);
        writeFile(TRANSFORMED_JAR, code);
    }

    @Override
    public byte[] getObjectGraph(AionAddress address) {
        logger.debug("[getObjectGraph] {} ", address);
        return readFile(OBJECT_GRAPH);
    }

    @Override
    public void putObjectGraph(AionAddress address, byte[] objectGraph) {
        logger.debug("[putObjectGraph] {} len={}", address, objectGraph.length);
        writeFile(OBJECT_GRAPH, objectGraph);
    }

    @Override
    public void putStorage(AionAddress address, byte[] key, byte[] value) {
        logger.debug("[putStorage] key={} value={}", Bytes.toHexString(key), Bytes.toHexString(value));
        try {
            proxy.setValue(key, value);
        } catch (IOException e) {
            logger.error("[putStorage] {}", e.getMessage());
        }
    }

    @Override
    public void removeStorage(AionAddress address, byte[] key) {
        logger.debug("[removeStorage] key={}", Bytes.toHexString(key));
        try {
            proxy.setValue(key, null);
        } catch (IOException e) {
            logger.error("[removeStorage] {}", e.getMessage());
        }
    }

    @Override
    public byte[] getStorage(AionAddress address, byte[] key) {
        try {
            byte[] value = proxy.getValue(key);
            logger.debug("[getStorage] key={} value={}", Bytes.toHexString(key), Bytes.toHexString(value));
            return value;
        } catch (IOException e) {
            logger.error("[getStorage] {}", e.getMessage());
            return null;
        }
    }

    @Override
    public void deleteAccount(AionAddress address) {
        logger.debug("[deleteStorage] {}", address);
        throw new RuntimeException("not implemented");
    }

    @Override
    public BigInteger getBalance(AionAddress address) {
        try {
            BigInteger balance = proxy.getBalance(address.toAddress());
            logger.debug("[getBalance] {} balance={}", address, balance);
            return balance;
        } catch (IOException e) {
            logger.error("[getBalance] {}", e.getMessage());
            return BigInteger.ZERO;
        }
    }

    @Override
    public void adjustBalance(AionAddress address, BigInteger amount) {
        logger.debug("[adjustBalance] {} amount={}", address, amount);
        // just ignore this
    }

    @Override
    public BigInteger getNonce(AionAddress address) {
        logger.debug("[getNonce] {}", address);
        throw new RuntimeException("not implemented");
    }

    @Override
    public void incrementNonce(AionAddress address) {
        logger.debug("[incrementNonce] {}", address);
        // just ignore this
    }

    @Override
    public void refundAccount(AionAddress address, BigInteger refund) {
        logger.debug("[refundAccount] {} refund={}", address, refund);
        throw new RuntimeException("not implemented");
    }

    @Override
    public byte[] getBlockHashByNumber(long blockNumber) {
        logger.debug("[getBlockHashByNumber] blockNumber={}", blockNumber);
        throw new RuntimeException("not implemented");
    }

    @Override
    public boolean accountNonceEquals(AionAddress address, BigInteger nonce) {
        logger.debug("[accountNonceEquals] {} nonce={}", address, nonce);
        return true;
    }

    @Override
    public boolean accountBalanceIsAtLeast(AionAddress address, BigInteger amount) {
        logger.debug("[accountBalanceIsAtLeast] {} amount={}", address, amount);
        return true;
    }

    @Override
    public boolean isValidEnergyLimitForCreate(long limit) {
        logger.debug("[isValidEnergyLimitForCreate] limit={}", limit);
        return true;
    }

    @Override
    public boolean isValidEnergyLimitForNonCreate(long limit) {
        logger.debug("[isValidEnergyLimitForNonCreate] limit={}", limit);
        return true;
    }

    @Override
    public boolean destinationAddressIsSafeForThisVM(AionAddress address) {
        logger.debug("[destinationAddressIsSafeForThisVM] {}", address);
        throw new RuntimeException("not implemented");
    }

    @Override
    public long getBlockNumber() {
        logger.debug("[getBlockNumber] ret={}", blockNumber);
        return blockNumber;
    }

    @Override
    public long getBlockTimestamp() {
        logger.debug("[getBlockTimestamp] ret={}", blockTimestamp);
        return blockTimestamp;
    }

    @Override
    public long getBlockEnergyLimit() {
        logger.debug("[getBlockEnergyLimit] ret={}", 0);
        return 0;
    }

    @Override
    public BigInteger getBlockDifficulty() {
        logger.debug("[getBlockDifficulty] ret={}", 0);
        return BigInteger.ZERO;
    }

    @Override
    public AionAddress getMinerAddress() {
        byte[] arr = new byte[AionAddress.LENGTH];
        Arrays.fill(arr, (byte) 0xaa);
        arr[0] = (byte) 0x0; // EOA
        AionAddress miner = new AionAddress(arr);
        logger.debug("[getMinerAddress] ret={}", miner);
        return miner;
    }
}
