// Copyright 2017 The Kubernetes Authors.
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

package args

import "net"

var builder = &holderBuilder{holder: Holder}

// Used to build argument holder structure. It is private to make sure that only 1 instance can be created
// that modifies singleton instance of argument holder.
type holderBuilder struct {
	holder *holder
}

// SetInsecurePort 'insecure-port' argument of Dashboard binary.
func (self *holderBuilder) SetInsecurePort(port int) *holderBuilder {
	self.holder.insecurePort = port
	return self
}

// SetInsecureBindAddress 'insecure-bind-address' argument of Dashboard binary.
func (self *holderBuilder) SetInsecureBindAddress(ip net.IP) *holderBuilder {
	self.holder.insecureBindAddress = ip
	return self
}

// GetHolderBuilder returns singleton instance of argument holder builder.
func GetHolderBuilder() *holderBuilder {
	return builder
}
