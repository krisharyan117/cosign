// Copyright 2021 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package blob

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type UnrecognizedSchemeError struct {
	Scheme string
}

func (e *UnrecognizedSchemeError) Error() string {
	return fmt.Sprintf("loading URL: unrecognized scheme: %s", e.Scheme)
}

func LoadFileOrURL(fileRef string) ([]byte, error) {
	var raw []byte
	var err error
	parts := strings.SplitAfterN(fileRef, "://", 2)
	if len(parts) == 2 {
		scheme := parts[0]
		switch scheme {
		case "http://":
			fallthrough
		case "https://":
			// #nosec G107
			resp, err := http.Get(fileRef)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			raw, err = io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
		case "env://":
			envVar := parts[1]
			// Most of Cosign should use `env.LookupEnv` (see #2236) to restrict us to known environment variables
			// (usually `$COSIGN_*`). However, in this case, `envVar` is user-provided and not one of the allow-listed
			// env vars.
			value, found := os.LookupEnv(envVar) //nolint:forbidigo
			if !found {
				return nil, fmt.Errorf("loading URL: env var $%s not found", envVar)
			}
			raw = []byte(value)
		default:
			return nil, &UnrecognizedSchemeError{Scheme: scheme}
		}
	} else {
		raw, err = os.ReadFile(filepath.Clean(fileRef))
		if err != nil {
			return nil, err
		}
	}
	return raw, nil
}
func LoadFileOrURLUsingReader(fileRef string) (io.Reader, error) {
	parts := strings.SplitAfterN(fileRef, "://", 2)
	if len(parts) == 2 {
		scheme := parts[0]
		switch scheme {
		case "http://":
			fallthrough
		case "https://":
			// #nosec G107
			resp, err := http.Get(fileRef)
			if err != nil {
				return nil, err
			}
			return resp.Body, nil
		case "env://":
			envVar := parts[1]
			// Most of Cosign should use `env.LookupEnv` (see #2236) to restrict us to known environment variables
			// (usually `$COSIGN_*`). However, in this case, `envVar` is user-provided and not one of the allow-listed
			// env vars.
			value, found := os.LookupEnv(envVar) //nolint:forbidigo
			if !found {
				return nil, fmt.Errorf("loading URL: env var $%s not found", envVar)
			}
			return strings.NewReader(value), nil
		default:
			return nil, &UnrecognizedSchemeError{Scheme: scheme}
		}
	}

		file,err := os.Open(filepath.Clean(fileRef))
		if err != nil {
			return nil, err
		}
	
	return file, nil
}
func LoadFileOrURLWithChecksum(fileRef string, checksum string) ([]byte, error) {
	checksumParts := strings.Split(checksum, ":")
	if len(checksumParts) >= 3 {
		return nil, fmt.Errorf("wrong checksum input format, must have at most 1 colon: %s", checksum)
	}

	checksumAlgo := sha256.New()
	checksumValue := checksumParts[len(checksumParts)-1]
	if len(checksumParts) == 2 {
		switch checksumParts[0] {
		case "sha256": // the default set above
		case "sha512":
			checksumAlgo = sha512.New()
		default:
			return nil, fmt.Errorf("unsupported checksum algorithm: %s", checksumParts[0])
		}
	}

	fileContent, err := LoadFileOrURL(fileRef)
	if err != nil {
		return nil, err
	}

	checksumAlgo.Write(fileContent)
	computedChecksum := hex.EncodeToString(checksumAlgo.Sum(nil))
	if computedChecksum != checksumValue {
		return nil, fmt.Errorf("incorrect checksum for file %s: expected %s but got %s", fileRef, checksumValue, computedChecksum)
	}

	return fileContent, nil
}
