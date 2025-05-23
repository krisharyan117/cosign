//
// Copyright 2023 The Sigstore Authors.
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

package remote

import (
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Referrers fetches references using registry options.
func Referrers(d name.Digest, artifactType string, opts ...Option) (*v1.IndexManifest, error) {
	o := makeOptions(name.Repository{}, opts...)
	rOpt := o.ROpt
	if artifactType != "" {
		rOpt = append(rOpt, remote.WithFilter("artifactType", artifactType))
	}
	idx, err := remote.Referrers(d, rOpt...)
	if err != nil {
		return nil, err
	}
	return idx.IndexManifest()
}
