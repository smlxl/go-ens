// Copyright 2017-2023 Weald Technology Trading.
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

package ens

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/wealdtech/go-ens/v3/contracts/reverseresolver"
)

// ReverseResolver is the structure for the reverse resolver contract.
type ReverseResolver struct {
	Contract     *reverseresolver.Contract
	ContractAddr common.Address
	ChainId      ChainId
}

// NewReverseResolverFor creates a reverse resolver contract for the given address.
func NewReverseResolverFor(backend bind.ContractBackend, address common.Address, chainId ChainId) (*ReverseResolver, error) {
	registry, err := NewRegistry(backend, chainId)
	if err != nil {
		return nil, err
	}

	// Now fetch the resolver.
	n := getRegistryAddress(chainId)
	domain := fmt.Sprintf("%x.%s", address.Bytes(), n)
	contractAddress, err := registry.ResolverAddress(domain)
	if err != nil {
		return nil, err
	}
	return NewReverseResolverAt(backend, contractAddress, chainId)
}

// NewReverseResolver obtains the reverse resolver.
func NewReverseResolver(backend bind.ContractBackend, chainId ChainId) (*ReverseResolver, error) {
	reverseRegistrar, err := NewReverseRegistrar(backend, chainId)
	if err != nil {
		return nil, err
	}

	// Now fetch the default resolver.
	address, err := reverseRegistrar.DefaultResolverAddress()
	if err != nil {
		return nil, err
	}

	return NewReverseResolverAt(backend, address, chainId)
}

// NewReverseResolverAt obtains the reverse resolver at a given address.
func NewReverseResolverAt(backend bind.ContractBackend, address common.Address, chainId ChainId) (*ReverseResolver, error) {
	// Instantiate the reverse registrar contract.
	contract, err := reverseresolver.NewContract(address, backend)
	if err != nil {
		return nil, err
	}

	// Ensure the contract is a resolver.
	n := getRegistryAddress(chainId)
	name := fmt.Sprintf("0.%s", n)
	nameHash, err := NameHash(name)
	if err != nil {
		return nil, err
	}
	_, err = contract.Name(nil, nameHash)
	if err != nil && err.Error() == "no contract code at given address" {
		return nil, fmt.Errorf("not a resolver")
	}

	return &ReverseResolver{
		Contract:     contract,
		ContractAddr: address,
		ChainId:      chainId,
	}, nil
}

// Name obtains the name for an address.
func (r *ReverseResolver) Name(address common.Address) (string, error) {
	ra := getRegistryAddress(r.ChainId)
	n := fmt.Sprintf("%s.%s", address.Hex()[2:], ra)
	nameHash, err := NameHash(n)
	if err != nil {
		return "", err
	}
	return r.Contract.Name(nil, nameHash)
}

// Format provides a string version of an address, reverse resolving it if possible.
func Format(backend bind.ContractBackend, address common.Address, chainId ChainId) string {
	result, err := ReverseResolve(backend, address, chainId)
	if err != nil {
		result = address.Hex()
	}
	return result
}

// ReverseResolve resolves an address in to an ENS name.
// This will return an error if the name is not found or otherwise 0.
func ReverseResolve(backend bind.ContractBackend, address common.Address, chainId ChainId) (string, error) {
	resolver, err := NewReverseResolverFor(backend, address, chainId)
	if err != nil {
		return "", err
	}

	// Resolve the name.
	name, err := resolver.Name(address)
	if err != nil {
		return "", err
	}
	if name == "" {
		err = errors.New("no resolution")
	}

	return name, err
}
