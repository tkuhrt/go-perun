# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] Despina - 2020-07-23 [:warning:]
Introduced a wire messaging abstraction. License changed to Apache 2.0.

### Added
- Wire messages are now sent and received over an abstract `wire.Bus`. It serves
  as a wire messaging abstraction for the `client` package using pub/sub
  semantics.
  - `wire.Msg`s are wrapped in `Envelope`s that have a sender and recipient.
  - Two implementations available:
    - `wire.LocalBus` for multiple clients running in the same program instance.
    - `wire/net.Bus` for wire connections over networks.
- Consistent use of `wallet.Address`es as map keys (`wallet.AddrKey`).
- Ordering to `wallet.Addresses`es to resolve ties in protocols.
- Contract validation to Ethereum backend.
- Consistent creation of PRNGs in tests (`pkg/test.Prng`).

### Changed
- License to Apache 2.0.
- The packages `peer`, `wire` and `pkg/io` were restructured:
  - Serialization code was moved from `wire` into `pkg/io`.
  - The `peer` package was merged into the `wire` package.
  - Networking-specific `wire` components were moved into `wire/net`.
  - The simple TCP/IP and Unix `Dialer` and `Listener` implementations were
    moved into `wire/net/simple`.
- The `ProposalHandler` and `UpdateHandler` interfaces' methods were renamed to
  explicitly name what they handle (`HandleProposal` and `HandleUpdate`).
- The keyvalue persister uses an improved data model and doesn't cache
  peer-channels any more.
- `Channel.Peers` now returns the full list of channel network peers, including
  the own `wire.Address`.

### Fixed
- A race in `client` synchronization.

### Removed
- The `net` package, as it didn't contain anything useful.

## [0.3.0] Charon - 2020-05-29 [:warning:]
Added persistence module to persist channel state data and handle client
shutdowns/restarts, as well as disconnects/reconnects.

### Added
- Persistence:
  - Persister, Restorer, ChannelIterator interfaces to allow for multiple
    persistence implementations.
    - sortedkv implementation provided (in-memory and LevelDB).
  - States and signatures are constantly persisted while channels progress.
  - Clients restore all saved channels on startup. State is synchronized with peers.
  - `Client.OnNewChannel` callback registration to deal with restored
    channels.
- Wallet interface for account unlocking abstraction.
  - Used during persistence to restore secret keys for signing.
  - Implemented for the Ethereum and simulated backend.
- Peer disconnect/reconnect handling.
- `Channel.UpdateBy` functional channel update method for better usability.

### Changed
- License changed to Apache 2.0.
- Replaced `Channel.ListenUpdates` and `Client.HandleChannelProposals` with
  `Client.Handle(ProposalHandler, UpdateHandler)` - a single common handler
  routine per client.
- Adapted client to new persistence layer and wallet.
- Made Ethereum interactions idempotent (increased safety).
- Moved subpackage `db` to `pkg/sortedkv`.
- Swapped Balance dimensions of type `channel.Allocation`.
- Random type generators in package `channel/test` now accept options to
  customize random data generation.
- Channels now get automatically closed when peers disconnect (and restored on reconnect).

### Fixed
- Ethereum backend: No funding transactions for zero own initial channel balances.

## [0.2.0] Belinda - 2020-03-23 [:warning:]
Added direct disputes and watcher for two-party ledger channels, much polishing
(refactors, bug fixes, documentation).

### Added
- Ledger state channel disputes and watcher.
- `channel.Adjudicator` interface and Ethereum implementation for registering
  channel states and withdrawing concluded channels.
- [Ethereum contracts](https://github.com/perun-network/contracts-eth) for disputes
- Public [Github wiki](https://github.com/perun-network/go-perun/wiki)
- [Godoc](https://pkg.go.dev/perun.network/go-perun)
- Changelog
- [TravisCI](https://travis-ci.org/perun-network)
- [goreportcard](https://goreportcard.com/report/github.com/perun-network/go-perun)
- [codeclimate](https://codeclimate.com/github/perun-network/go-perun)
- TCP and unix socket `peer.Dialer` and `Listener` implementations.
- `Eventually` tester in `pkg/test` to repeatedly run tests until they
  succeed.
- Concurrent testing tool in `pkg/test` to be able to call `require` in tests
  with multiple go routines.

### Changed
- `client.New` now needs a `Funder` and `Adjudicator`, instead of a `Settler`.
- `Serializable` renamed to `Serializer`.
- Unified backend imports.
- `pkg/io/test/bytewiseReader` to `iotest.OneByteReader`.
- Improved peer message handling mechanism.
- Consistent handling of `nil` arguments in exported functions.
- Many refactors to improve the overall code quality and documentation.
- Updated Ethereum contract bindings to newest version.

### Removed
- `wallet.Wallet` interface and `sim` backend implementation - it was never used.
- `ethereum` and `sim/wallet.NewAddressFromBytes` - only `wallet.DecodeAddress`
  should be used to create an `Address` from bytes.
- `channel/machine` Phase subscription logic.
- `channel.Settler` interface and backend implementations - replaced by `Adjudicator`.

### Fixed
- Reduced cyclomatic complexity of complex functions.
- Deadlock in two-party payment channel test.
- Ethereum backend test timeouts and instabilities.
- Many minor bug fixes, mainly concurrency issues in tests.

## [0.1.0] Ariel - 2019-12-20 [:warning:]
Initial release.

### Added
- Two-party ledger state channels.
- Cooperatively settling two-party ledger channels.
- Ethereum blockchain backend.
- Logrus logging backend.


## Legend
- <a name=":warning:">:warning:</a>: This release is not suited for practical use with real money.


[:warning:]: #:warning:

[Unreleased]: https://github.com/perun-network/go-perun/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/perun-network/go-perun/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/perun-network/go-perun/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/perun-network/go-perun/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/perun-network/go-perun/releases/v0.1.0
