# Change Log

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [2.2.12]
### Added
- support subscription manger instance

## [2.2.10]
### Changed
- move the check auth to the init func

## [2.2.9]
### Changed
- move the subdomain arg to be only used in the CLI interactive mode
- add the option to save the query result to a file in the run-query command

## [2.2.8]
### Added
- support `GCP` db for `windows`

## [2.2.7]
### Added
- support `GCP` db

## [2.1.51]
### Fixed
- update cli on exist from Main as well

## [2.1.4]
### Fixed
- add default team feature and update cli on exist

## [2.1.3]
### Fixed
- revert comment in code

## [2.1.2]
### Changed
- removing "closing connection"
- add "log"

## [2.1.1]
### Fixed
- wait for the connection of redis to be closed before running a new one.

## [2.1.0]
### Added
- Add log to which DB we are connecting when running the Postgres command.
- Adding CheckCertCmd for Ops team features
- Automatically move to the Master node when using Redis command

## [2.0.1]
### Added
- Add a new command to get all the log level for all apps.

## [2.0.0]
### Added
- Add a new sub-command to connect to a DB immediately without having to specify an app name.
- Create clear and understandable help commands for each of the commands.
### Changed
- Support CN40 landscape. 
- Improve change-target to handle flags as well 
- Improve change-target performance 
- Full completion for MAC - both commands and dynamic completion (apps, instances, orgs, and spaces)
- Change logs color 
- The token will now be copied to the clipboard instead of being printed out.
## [1.1.38]
### Fixed
- fix update flow

## [1.1.37]
### Changed
- revert cf change

## [1.1.36]
### Changed
- improve db query

## [1.1.35]
### fixed
- try to fix db connection on windows

## [1.1.34]
### fixed
- fix selectSpace

## [1.1.33]
### Changed
- improve db query

## [1.1.32]
### Changed
- change and improve

## [1.1.31]
### Updated
- change option name

## [1.1.30]
### Fixed
- order env by name
- improve log by level

## [1.1.29]
### Fixed
- improve handle connection

## [1.1.28]
### Fixed
- do not crash when app is changing 

## [1.1.27]
### Added
- support message count

## [1.1.26]
### Added
- support message count

## [1.1.25]
### Fixed
- fixed postgres connection issue

## [1.1.24]
### Fixed
- fixed postgres connection issue

## [1.1.23]
### Updated
- Order the SAA by status and tenantId 

## [1.1.22]
### Added
- Added the ability to enable ssh when trying to connect to DB

## [1.1.21]
### Fixed
- Fix timeout

## [1.1.20]
### Fixed
- Fix migNumbers

## [1.1.19]
### Added
- Added the ability to run query on one db for all landscapes (currently only for Ops only).

## [1.1.18]
### Added
- Added the ability to run query on db with the `query` command (currently only for Ops only).
- Added level option for `logs-recent` command.

## [1.1.17]
### Added
- Support `chunks` for SAA in batch mode.

## [1.1.16]
### Fixed
- fix the Ops teamFeatures

## [1.1.15]
### Added
- Add the teamFeatures feature.

## [1.1.14]
### Fixed
- Fix credentials getting in FF service

## [1.1.13]
### Fixed
- Fix the version comparison

## [1.1.12]
### Fixed
- Fix the auth

## [1.1.11]
### Fixed
- create team with the correct team name

## [1.1.10]
### Fixed
- Fix the auth

## [1.1.9]
### Added
- Add the migrations files to support new features

## [1.1.8]
### Fixed
- Fix the issue with the blue-green deployment support.

## [1.1.7]
### Updated
- Improve recent logs feature

## [1.1.6]
### Added
- Added the functionality for the message-queue service

### Updated
- improve the performance of the update landscape function

## [1.1.5]
### Added
- Added show credentials action for instance

## [1.1.4]
### Added
- Now we support Correlation ID for the logs-recent command
- Added a status command

### Fixed
- Fix the FF service issue.

## [1.1.3]
### Updated
- improve mutex lock to avoid map read-write conflict

## [1.1.2]
### Updated
- Add mutex lock to avoid map read-write conflict

## [1.1.1]
### Updated
- Move the error handling after the batch selection.

## [1.1.0]
### Added
- Add the ability to create a new key

## [1.0.14]
### Added
- Added a cache for the instances of the CF landscape
- Complete the Batch mode

## [1.0.13]
### Updated
- improve get credentials from the environment variables
- Support windows on autoUpdate file

### Added
- Added saas-registry capabilities

## [1.0.12]
### Fixed
- Fix the auto update feature

## [1.0.11]
### Updated
- Add files to a one zip

## [1.0.10]
### Updated
- Add files to a zip

## [1.0.9]
### Updated
- change the file name for the version file

## [1.0.8]
### Added
- Add the auto update feature to the artifactory

## [1.0.7]
### Added
- support auto update for the CLI

## [1.0.6]
### Added
- Add a latest version file to the repository

## [1.0.5]
### Added
- Add file to update the latest version

## [1.0.4]
### Added
- Add deployment to artifctory

## [1.0.3]
### Added
- Add the ability to create token based on subdomain.

## [1.0.2]
### Fix
- change file name.

## [1.0.1]
### Added
- Base version of the CLI
