package otr3

import "errors"

var (
	// GPG_ERR_NO_ERROR is matched to nil
	// GPG_ERR_ENOMEM doesn't make any sense in golang

	// ErrGPGUnusableSecretKey maps to GPG_ERR_UNUSABLE_SECKEY in libotr
	ErrGPGUnusableSecretKey = errors.New("GPG Error: Unusable secret key (54)")
	// ErrGPGInvalidValue maps to GPG_ERR_INV_VALUE in libotr
	ErrGPGInvalidValue = errors.New("GPG Error: Invalid value (55)")
	// ErrGPGConflict maps to GPG_ERR_CONFLICT in libotr
	ErrGPGConflict = errors.New("GPG Error: Conflict (70)")
	// ErrGPGEntityExist maps to GPG_ERR_EEXIST in libotr
	ErrGPGEntityExist = errors.New("GPG Error: Entity exist (32803)")
)