# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

- GitHub Actions CI/CD pipeline
- Daily security scanning with gosec

## [0.0.1] - 2024-09-18

### Added

- Initial Go port of uuidv47 C library
- Core functions: `Encode()`, `Decode()`, `Parse()`, `String()`, `NewRandomKey()`
- Exact compatibility with C implementation
- Comprehensive test suite
- Benchmarks (~9ns per operation)
- Example program

[Unreleased]: https://github.com/n2p5/uuid47/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/n2p5/uuid47/releases/tag/v0.0.1
