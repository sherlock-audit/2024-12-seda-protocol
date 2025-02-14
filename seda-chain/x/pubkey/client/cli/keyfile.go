package cli

import (
	"bufio"
	"crypto/aes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bgentry/speakeasy"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

func LoadOrGenerateSEDAKeys(cmd *cobra.Command, valAddr sdk.ValAddress) ([]types.IndexedPubKey, error) {
	serverCfg := server.GetServerContextFromCmd(cmd).Config

	encryptionKeyEnv := utils.ReadSEDAKeyEncryptionKeyFromEnv()
	useCustomEncryptionKey, err := cmd.Flags().GetBool(FlagEncryptionKey)
	if err != nil {
		return nil, err
	}

	if useCustomEncryptionKey && encryptionKeyEnv != "" {
		return nil, fmt.Errorf("both --%s flag and %s environment variable are set", FlagEncryptionKey, utils.SEDAKeyEncryptionKeyEnvVar)
	}

	encryptionKey := encryptionKeyEnv
	if useCustomEncryptionKey {
		customKey, err := speakeasy.FAsk(os.Stderr, "Enter the custom encryption key\n")
		if err != nil {
			return nil, err
		}
		confirmation, err := speakeasy.FAsk(os.Stderr, "Confirm the custom encryption key\n")
		if err != nil {
			return nil, err
		}
		if confirmation != customKey {
			return nil, fmt.Errorf("custom encryption key confirmation does not match")
		}

		customKeyBytes, err := base64.StdEncoding.DecodeString(customKey)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 encoded key: %w", err)
		}

		_, err = aes.NewCipher(customKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("invalid AES key: %w", err)
		}

		encryptionKey = customKey
	}

	var pks []types.IndexedPubKey
	keyFile, err := cmd.Flags().GetString(FlagKeyFile)
	if err != nil {
		return nil, err
	}

	if keyFile != "" {
		pks, err = utils.LoadSEDAPubKeys(keyFile, encryptionKey)
		if err != nil {
			return nil, err
		}
	} else {
		encryptionKey, err = getSEDAKeysEncryptionKey(cmd, encryptionKey)
		if err != nil {
			return nil, err
		}

		forceKeyFile, err := cmd.Flags().GetBool(FlagForceKeyFile)
		if err != nil {
			return nil, err
		}

		if utils.SEDAKeyFileExists(filepath.Dir(serverCfg.PrivValidatorKeyFile())) && !forceKeyFile {
			reader := bufio.NewReader(os.Stdin)
			overwrite, err := input.GetConfirmation("SEDA key file already exists, overwrite?", reader, os.Stderr)
			if err != nil {
				return nil, err
			}

			forceKeyFile = overwrite
		}

		pks, err = utils.GenerateSEDAKeys(valAddr, filepath.Dir(serverCfg.PrivValidatorKeyFile()), encryptionKey, forceKeyFile)
		if err != nil {
			return nil, err
		}
	}

	return pks, nil
}

func getSEDAKeysEncryptionKey(cmd *cobra.Command, encryptionKey string) (string, error) {
	if encryptionKey != "" {
		return encryptionKey, nil
	}

	noEncryptionFlag, err := cmd.Flags().GetBool(FlagNoEncryption)
	if err != nil {
		return "", err
	}

	_, allowUnencrypted := os.LookupEnv(utils.EnvAllowUnencryptedSedaKeys)
	if noEncryptionFlag || allowUnencrypted {
		return "", nil
	}

	encryptionKey, err = utils.GenerateSEDAKeyEncryptionKey()
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(os.Stdin)
	confirmation, err := input.GetConfirmation(fmt.Sprintf("\n**Important** take note of this encryption key.\nIt is required as an env variable (%s) when running the node.\n\n%s\n", utils.SEDAKeyEncryptionKeyEnvVar, encryptionKey), reader, os.Stderr)
	if err != nil {
		return "", err
	}
	if !confirmation {
		return "", fmt.Errorf("user did not confirm the generated encryption key")
	}

	return encryptionKey, nil
}
