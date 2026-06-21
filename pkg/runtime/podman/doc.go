// Package podman is the native Podman implementation of
// runtime.ContainerRuntime, built on Podman's Go bindings
// (github.com/containers/podman/v5/pkg/bindings).
//
// Phase 3a (this code) ships the scaffolding only: a Runtime that
// satisfies the interface but reports runtime.ErrUnsupported for every
// operation, plus the config/env plumbing that lets a user select it.
// The bindings dependency and the real connection setup land in Phase 3b
// alongside the first implemented method group, so the heavy dependency
// tree is not pulled in until it is actually used. See
// docs/adr/0005-podman-native-backend.md.
package podman
