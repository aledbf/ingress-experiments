#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null)}

${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
  github.com/aledbf/ingress-experiments/internal/apis github.com/aledbf/ingress-experiments/internal/apis \
  configuration:v1alpha1 \
  --output-base "$(dirname ${BASH_SOURCE})/../../../.." \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate/boilerplate.go.txt

${GOPATH}/bin/conversion-gen || go build -o "${GOPATH}/bin/conversion-gen" "${CODEGEN_PKG}/cmd/conversion-gen"

${GOPATH}/bin/conversion-gen --skip-unsafe=true \
  --input-dirs github.com/aledbf/ingress-experiments/internal/apis/configuration/v1alpha1 \
  --output-file-base=zz_generated.conversion \
  --go-header-file "hack/boilerplate/boilerplate.go.txt"
