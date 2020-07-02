# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.2] - 2020-07-01
### Added
- This changelog.
- Version number.

## [0.1.1] - 2020-06-16
### Changed
- Better command documentation.
- Replace [urfave/cli](https://github.com/urfave/cli) with
    [spf13/cobra](https://github.com/spf13/cobra) because it appears that with
    `spf13/cobra` it is easier to make the desired documentation changes.

## [0.1.0] - 2020-06-13
### Changed
- The `sync` subcommand will now read from Stdin by default.  Otherwise you must
    specify a file to read from like so:

        repos -f path/to/config.repo sync

    This addresses issue #2.
- The `sync` subcommand no longer takes any arguments.

### Removed
- The `-c, --config` options are removed.  This removes the strict configuration structure imposed before.

## [0.0.0] - 2020-02-28
### Changed
- The `fetch` subcommand is renamed to `sync`. This is to avoid confusion with
    the `git fetch` behavior.
- For the `sync` subcommand, if the worktree is unclean a fetch is performed,
    otherwise a pull is attempted.
- Both sync and parse take channel of errors this is used to send errors that
    occur that do not need to stop the process. For example, if a repo cannot
    sync the error is sent through the channel but we continue to sync the next
    repo.

## [0.0.0] - 2019-10-15
### Added
- Initial commit.  The start of this project.
