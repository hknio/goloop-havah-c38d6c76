package io.havah.test.cases;

import foundation.icon.icx.IconService;
import foundation.icon.icx.KeyWallet;
import foundation.icon.icx.Wallet;
import foundation.icon.icx.data.Bytes;
import foundation.icon.icx.data.TransactionResult;
import foundation.icon.test.common.TestBase;
import foundation.icon.test.common.TransactionHandler;
import io.havah.test.common.Constants;
import io.havah.test.common.Utils;
import io.havah.test.score.SustainableFundScore;
import io.havah.test.score.VaultScore;
import jdk.jshell.execution.Util;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Order;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import java.math.BigInteger;

import static foundation.icon.test.common.Env.LOG;
import static org.junit.jupiter.api.Assertions.assertEquals;

@Tag(Constants.TAG_HAVAH)
public class VaultTest extends TestBase {
    private static TransactionHandler txHandler;
    private static KeyWallet[] wallets;
    private static VaultScore vaultScore;

    private static Wallet ownerWallet;

    @BeforeAll
    static void setup() throws Exception {
        txHandler = Utils.getTxHandler();
        wallets = new KeyWallet[3];
        BigInteger amount = ICX.multiply(BigInteger.valueOf(300));
        for (int i = 0; i < wallets.length; i++) {
            wallets[i] = KeyWallet.create();
            txHandler.transfer(wallets[i].getAddress(), amount);
        }
        for (KeyWallet wallet : wallets) {
            ensureIcxBalance(txHandler, wallet.getAddress(), BigInteger.ZERO, amount);
        }

        vaultScore = new VaultScore(txHandler);

        ownerWallet = Utils.getGovernor();
        txHandler.transfer(ownerWallet.getAddress(), amount);
        ensureIcxBalance(txHandler, ownerWallet.getAddress(), BigInteger.ZERO, amount);
    }

    @Test
    void startVault() throws Exception {
        LOG.infoEntering("call", "setAccounts()");
        VaultScore.VestingAccount[] accounts = {
                new VaultScore.VestingAccount(wallets[0].getAddress(), BigInteger.valueOf(5000)),
                new VaultScore.VestingAccount(wallets[1].getAddress(), BigInteger.valueOf(3000)),
                new VaultScore.VestingAccount(wallets[2].getAddress(), BigInteger.valueOf(2000))
        };

        BigInteger curHeight = Utils.getHeight();

        BigInteger[] heights = {
                curHeight.add(BigInteger.valueOf(10)),
                curHeight.add(BigInteger.valueOf(15)),
                curHeight.add(BigInteger.valueOf(30))
        };
        assertSuccess(vaultScore.setAccounts(ownerWallet, accounts, heights));

        Bytes txHash = txHandler.transfer(wallets[0], vaultScore.getAddress(), ICX.multiply(BigInteger.valueOf(100)));
        assertFailure(txHandler.getResult(txHash));
        LOG.infoExiting();

        LOG.infoEntering("call", "claim() [0] : " + wallets[0].getAddress());
        LOG.info("getClaimableAmount before claim : " + vaultScore.getClaimableAmount(wallets[0].getAddress()));
        LOG.info("wallet balance before claim : " + txHandler.getBalance(wallets[0].getAddress()));
        assertSuccess(vaultScore.claim(wallets[0]));
        LOG.info("getClaimableAmount after claim : " + vaultScore.getClaimableAmount(wallets[0].getAddress()));
        LOG.info("wallet balance after claim : " + txHandler.getBalance(wallets[0].getAddress()));
        LOG.info("getBalanceOf : " + vaultScore.getBalanceOf(wallets[0].getAddress()));
        LOG.info("vault balance : " + txHandler.getBalance(vaultScore.getAddress()));
        LOG.infoExiting();

        Utils.waitUtil(BigInteger.valueOf(25));

        LOG.infoEntering("call", "claim() [0] : " + wallets[0].getAddress());
        LOG.info("getClaimableAmount before claim : " + vaultScore.getClaimableAmount(wallets[0].getAddress()));
        LOG.info("wallet balance before claim : " + txHandler.getBalance(wallets[0].getAddress()));
        assertSuccess(vaultScore.claim(wallets[0]));
        LOG.info("getClaimableAmount after claim : " + vaultScore.getClaimableAmount(wallets[0].getAddress()));
        LOG.info("wallet balance after claim : " + txHandler.getBalance(wallets[0].getAddress()));
        LOG.info("getBalanceOf : " + vaultScore.getBalanceOf(wallets[0].getAddress()));
        LOG.info("vault balance : " + txHandler.getBalance(vaultScore.getAddress()));
        LOG.infoExiting();

        Utils.waitUtil(BigInteger.valueOf(35));

        LOG.infoEntering("call", "claim() [0] : " + wallets[0].getAddress());
        LOG.info("getClaimableAmount before claim : " + vaultScore.getClaimableAmount(wallets[0].getAddress()));
        LOG.info("wallet balance before claim : " + txHandler.getBalance(wallets[0].getAddress()));
        assertSuccess(vaultScore.claim(wallets[0]));
        LOG.info("getClaimableAmount after claim : " + vaultScore.getClaimableAmount(wallets[0].getAddress()));
        LOG.info("wallet balance after claim : " + txHandler.getBalance(wallets[0].getAddress()));
        LOG.info("getBalanceOf : " + vaultScore.getBalanceOf(wallets[0].getAddress()));
        LOG.info("vault balance : " + txHandler.getBalance(vaultScore.getAddress()));
        LOG.infoExiting();

        LOG.infoEntering("call", "claim() [1] : " + wallets[1].getAddress());
        LOG.info("getClaimableAmount before claim : " + vaultScore.getClaimableAmount(wallets[1].getAddress()));
        LOG.info("wallet balance before claim : " + txHandler.getBalance(wallets[1].getAddress()));
        assertSuccess(vaultScore.claim(wallets[1]));
        LOG.info("getClaimableAmount after claim : " + vaultScore.getClaimableAmount(wallets[1].getAddress()));
        LOG.info("wallet balance after claim : " + txHandler.getBalance(wallets[1].getAddress()));
        LOG.info("getBalanceOf : " + vaultScore.getBalanceOf(wallets[1].getAddress()));
        LOG.info("vault balance : " + txHandler.getBalance(vaultScore.getAddress()));
        LOG.infoExiting();

        Utils.waitUtil(BigInteger.valueOf(45));

        LOG.infoEntering("call", "claim() [0] : " + wallets[0].getAddress());
        LOG.info("getClaimableAmount before claim : " + vaultScore.getClaimableAmount(wallets[0].getAddress()));
        LOG.info("wallet balance before claim : " + txHandler.getBalance(wallets[0].getAddress()));
        assertSuccess(vaultScore.claim(wallets[0]));
        LOG.info("getClaimableAmount after claim : " + vaultScore.getClaimableAmount(wallets[0].getAddress()));
        LOG.info("wallet balance after claim : " + txHandler.getBalance(wallets[0].getAddress()));
        LOG.info("getBalanceOf : " + vaultScore.getBalanceOf(wallets[0].getAddress()));
        LOG.info("vault balance : " + txHandler.getBalance(vaultScore.getAddress()));
        LOG.infoExiting();

        LOG.infoEntering("call", "claim() [1] : " + wallets[1].getAddress());
        LOG.info("getClaimableAmount before claim : " + vaultScore.getClaimableAmount(wallets[1].getAddress()));
        LOG.info("wallet balance before claim : " + txHandler.getBalance(wallets[1].getAddress()));
        assertSuccess(vaultScore.claim(wallets[1]));
        LOG.info("getClaimableAmount after claim : " + vaultScore.getClaimableAmount(wallets[1].getAddress()));
        LOG.info("wallet balance after claim : " + txHandler.getBalance(wallets[1].getAddress()));
        LOG.info("getBalanceOf : " + vaultScore.getBalanceOf(wallets[1].getAddress()));
        LOG.info("vault balance : " + txHandler.getBalance(vaultScore.getAddress()));
        LOG.infoExiting();

        Utils.waitUtil(BigInteger.valueOf(55));

        LOG.infoEntering("call", "claim() [0] : " + wallets[0].getAddress());
        LOG.info("getClaimableAmount before claim : " + vaultScore.getClaimableAmount(wallets[0].getAddress()));
        LOG.info("wallet balance before claim : " + txHandler.getBalance(wallets[0].getAddress()));
        assertSuccess(vaultScore.claim(wallets[0]));
        LOG.info("getClaimableAmount after claim : " + vaultScore.getClaimableAmount(wallets[0].getAddress()));
        LOG.info("wallet balance after claim : " + txHandler.getBalance(wallets[0].getAddress()));
        LOG.info("getBalanceOf : " + vaultScore.getBalanceOf(wallets[0].getAddress()));
        LOG.info("vault balance : " + txHandler.getBalance(vaultScore.getAddress()));
        LOG.infoExiting();

        LOG.infoEntering("call", "claim() [2] : " + wallets[2].getAddress());
        LOG.info("getClaimableAmount before claim : " + vaultScore.getClaimableAmount(wallets[2].getAddress()));
        LOG.info("wallet balance before claim : " + txHandler.getBalance(wallets[2].getAddress()));
        assertSuccess(vaultScore.claim(wallets[2]));
        LOG.info("getClaimableAmount after claim : " + vaultScore.getClaimableAmount(wallets[2].getAddress()));
        LOG.info("wallet balance after claim : " + txHandler.getBalance(wallets[2].getAddress()));
        LOG.info("getBalanceOf : " + vaultScore.getBalanceOf(wallets[2].getAddress()));
        LOG.info("vault balance : " + txHandler.getBalance(vaultScore.getAddress()));
        LOG.infoExiting();
    }
}
