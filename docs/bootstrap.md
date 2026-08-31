# Repository and authority facts

- The repository was bootstrapped directly on `main` in commit
  `499e141` before feature work began.
- All implementation and contract changes after that bootstrap are developed
  on `feature/language-delta-forge-v1` and are intended to land through one
  pull request.
- This product is an independent read-only consumer. It has zero repository
  write authority over protected Gooo core, zero merge authority, and zero
  adoption authority.
- Local test execution is intentionally zero. GitHub Actions is the required
  compile, build, test, and conformance authority.
- The root README is excluded from inventory counts.
