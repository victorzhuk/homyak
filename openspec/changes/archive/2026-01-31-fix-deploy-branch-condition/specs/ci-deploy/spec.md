## MODIFIED Requirements

### Requirement: Deploy workflow triggers on default branch

The deploy workflow SHALL trigger when the build workflow completes successfully on the `master` branch.

#### Scenario: Deploy runs after successful build on master
- **WHEN** the build workflow completes with success on `master` branch
- **THEN** the deploy workflow SHALL execute (not be skipped)

#### Scenario: Deploy skipped for non-master branches
- **WHEN** the build workflow completes on a branch other than `master`
- **THEN** the deploy workflow SHALL be skipped
