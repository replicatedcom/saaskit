package crypto

import (
	"bufio"
	"bytes"
	"regexp"

	"github.com/pkg/errors"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
)

type PGPKeyPair struct {
	PrivateKeyText string
	PublicKeyText  string
}

func GeneratePGPKeyPair(name, comment, email string) (*PGPKeyPair, error) {
	name = makeSafe(name)
	comment = makeSafe(comment)
	email = makeSafe(email)

	// ent type is *openpgp.Entity
	ent, err := openpgp.NewEntity(name, comment, email, nil)
	if err != nil {
		return nil, errors.Wrap(err, "create new entity")
	}

	for _, id := range ent.Identities {
		err := id.SelfSignature.SignUserId(id.UserId.Id, ent.PrimaryKey, ent.PrivateKey, nil)
		if err != nil {
			return nil, errors.Wrap(err, "sign user ID")
		}
	}

	for _, subkey := range ent.Subkeys {
		err := subkey.Sig.SignKey(subkey.PublicKey, ent.PrivateKey, nil)
		if err != nil {
			return nil, errors.Wrap(err, "sign key")
		}
	}

	var pubBuff, prvBuff bytes.Buffer
	pubWriter := bufio.NewWriter(&pubBuff)
	prvWriter := bufio.NewWriter(&prvBuff)

	prvEncoder, err := armor.Encode(prvWriter, openpgp.PrivateKeyType, nil)
	if err != nil {
		return nil, errors.Wrap(err, "encode secret key")
	}
	defer prvEncoder.Close()
	if err := ent.SerializePrivate(prvEncoder, nil); err != nil {
		return nil, errors.Wrap(err, "serialize secret key")
	}

	pubEncoder, err := armor.Encode(pubWriter, openpgp.PublicKeyType, nil)
	if err != nil {
		return nil, errors.Wrap(err, "encode public key")
	}
	defer pubEncoder.Close()

	if err := ent.Serialize(pubEncoder); err != nil {
		return nil, errors.Wrap(err, "serialize public key")
	}

	_ = pubWriter.Flush()
	_ = prvWriter.Flush()
	keyPair := &PGPKeyPair{
		PrivateKeyText: prvBuff.String(),
		PublicKeyText:  pubBuff.String(),
	}

	return keyPair, nil
}

// From the openpgp package source:
// NewUserId returns a UserId or nil if any of the arguments contain invalid
// characters. The invalid characters are '\x00', '(', ')', '<' and '>'
var safeRe = regexp.MustCompile("[\x00()<>]")

func makeSafe(s string) string {
	return safeRe.ReplaceAllString(s, "-")
}
