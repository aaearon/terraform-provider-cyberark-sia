<!--
SYNC IMPACT REPORT
==================
Version Change: [NEW] → 1.0.0
Created: 2025-10-15
Type: Initial constitution creation

Principles Added:
- I. Code Quality and SOLID Principles
- II. Maintainability First
- III. LLM-Driven Development
- IV. Light-Touch Testing
- V. Documentation as Code

Sections Added:
- Core Principles (5 principles)
- Development Standards
- Quality Gates
- Governance

Templates Status:
✅ .specify/templates/spec-template.md - Reviewed, no updates needed
✅ .specify/templates/plan-template.md - Reviewed, Constitution Check section aligns
✅ .specify/templates/tasks-template.md - Reviewed, testing approach aligns
⚠ .specify/templates/commands/*.md - Manual review recommended for consistency

Rationale for v1.0.0:
- Initial constitution establishing foundational governance
- Based on authoritative sources (Martin Fowler, Robert C. Martin, HashiCorp documentation)
- All principles sourced from documented best practices
- No assumptions made without factual basis

Follow-up TODOs:
- None: All placeholders filled with concrete values
-->

# Terraform Provider CyberArk SIA Constitution

## Core Principles

### I. Code Quality and SOLID Principles

**Source**: Robert C. Martin (Uncle Bob), "Clean Code" and SOLID design principles

Code MUST adhere to SOLID principles as the foundation for maintainable software:
- **Single Responsibility**: Each module, class, or function has one reason to change
- **Open/Closed**: Software entities are open for extension, closed for modification
- **Liskov Substitution**: Derived types must be substitutable for their base types
- **Interface Segregation**: Clients should not depend on interfaces they don't use
- **Dependency Inversion**: Depend on abstractions, not concretions

**Rationale**: SOLID principles establish practices for developing software with
considerations for maintaining and extending it as the project grows. These are
time-tested foundations recognized across the industry for object-oriented design.

**Application to Terraform Providers**:
- Provider resources and data sources MUST have single, well-defined purposes
- Schema definitions MUST be extensible without breaking existing configurations
- Resource implementations MUST be independently testable and composable

### II. Maintainability First

**Source**: Martin Fowler, "Refactoring"; HashiCorp Terraform Plugin Framework documentation

Code MUST prioritize readability and maintainability over cleverness:
- Clear, descriptive naming over brevity (no abbreviations unless universally recognized)
- Explicit error handling with actionable error messages
- Logical code organization following established project structure patterns
- Comments MUST explain "why", not "what" (the code itself explains "what")

**Go-Specific Standards**:
- Follow official Go Code Review Comments and Effective Go guidelines
- Use `gofmt` and `golangci-lint` for consistent formatting
- Leverage Go's built-in error handling patterns (return errors, don't panic)

**Terraform Provider Standards**:
- Follow HashiCorp's Terraform Plugin Framework conventions and patterns
- Resource CRUD operations MUST have clear separation of concerns
- Schema definitions MUST include descriptions for all attributes
- Validators and plan modifiers MUST be reusable and well-documented

**Rationale**: Code is read far more often than it is written. Maintainability
reduces technical debt, accelerates onboarding, and enables confident refactoring.

### III. LLM-Driven Development

**Source**: 2025 Prompt Engineering best practices (Lakera, Mirascope, Palantir);
GitHub Copilot and AI coding assistant documentation

Development leveraging LLMs MUST follow these practices:

**Clear Context Provision**:
- Provide LLMs with relevant context: existing code structure, related files, requirements
- Remove irrelevant context that could confuse the model
- Use structured prompts with clear instructions on desired output format

**Iterative Refinement**:
- Treat initial LLM output as a first draft requiring human review and refinement
- Iterate with specific feedback and additional constraints
- Verify LLM-generated code meets project standards before committing

**Version and Document**:
- Document any LLM interactions that significantly influence architecture decisions
- Version prompt templates used for code generation (if applicable)
- Human developers remain accountable for all committed code

**Structured Prompting**:
- Use role-based prompting when appropriate ("As a Go developer familiar with
  Terraform providers...")
- Break complex tasks into steps using chain-of-thought prompting
- Specify output format, language, and constraints explicitly

**Rationale**: LLMs are powerful productivity tools when used with discipline.
Structured interaction with LLMs produces higher quality, more maintainable code
while maintaining human oversight and accountability.

### IV. Light-Touch Testing

**Source**: Martin Fowler, "The Practical Test Pyramid"; Terraform Plugin Framework
testing documentation

Testing MUST follow the test pyramid principle with pragmatic application:

**Test Distribution** (Source: Martin Fowler, martinfowler.com/articles/practical-test-pyramid.html):
- **Many**: Unit tests - fast, isolated, testing individual functions/methods
- **Some**: Integration tests - testing contracts between components
- **Few**: Acceptance tests - testing end-to-end workflows

**Testing Requirements**:
- Write tests with different granularity levels
- The higher the level, the fewer tests needed
- Tests MUST run quickly and reliably
- Tests MUST only fail when there's a real problem
- Focus on value over coverage percentages

**Terraform Provider Testing**:
- Unit tests for validators, plan modifiers, and utility functions
- Acceptance tests for critical resource CRUD operations using terraform-plugin-testing
- Test state upgrades and schema migrations thoroughly
- Mock external API calls in unit tests; use real APIs sparingly in acceptance tests

**What NOT to Test**:
- Avoid test ice-cream cone anti-pattern (too many slow, brittle end-to-end tests)
- Don't test framework behavior (trust HashiCorp's testing of the framework itself)
- Don't aim for 100% coverage if tests provide no value

**Rationale**: The test pyramid ensures a healthy, fast, maintainable test suite.
Light-touch testing focuses effort on high-value tests that catch real issues
without creating maintenance burden.

### V. Documentation as Code

Documentation MUST be:
- **Continuous**: Updated as part of each task (not an afterthought)
- **Version-Controlled**: Treated as first-class code artifacts
- **Actionable**: Providing working examples and clear procedures
- **Current**: Never considered "done" - evolves with the codebase

**Required Documentation**:
- Resource and data source examples in Terraform configuration syntax
- Schema attribute descriptions in provider code
- Architecture decisions when non-obvious choices are made
- Setup and development quickstart guides

**Rationale**: Documentation drift is a primary source of developer friction.
Treating docs as code with the same quality standards ensures they remain useful.

## Development Standards

### Git Workflow

**MUST follow feature branch workflow**:
- All work happens in feature branches (`<issue-number>-<descriptive-name>`)
- Branches MUST be created from main/master
- Pull requests required for all changes to main/master
- Commits MUST be atomic and have descriptive messages

**Commit Message Format**:
- Follow Conventional Commits: `type(scope): description`
- Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`
- Example: `feat(resources): add role assignment resource`

### Code Review Standards

Pull requests MUST:
- Include a clear description of changes and rationale
- Reference related issues/specifications
- Pass all automated checks (linting, tests, builds)
- Receive at least one approval before merging

Reviewers MUST verify:
- Compliance with constitution principles
- Test coverage for new functionality
- Documentation updates included
- No unnecessary complexity introduced

### Tooling Requirements

**MUST use** (per user's CLAUDE.md):
- `fdfind` for file searching (not `find`)
- `rg` (ripgrep) for text/string searching (not `grep`)
- `ast-grep` for code structure searches
- `fzf` for selecting from multiple results
- `jq` for JSON interaction
- `yq` for YAML/XML interaction
- `gh` for GitHub interaction

**Go Development**:
- `go fmt` for formatting
- `golangci-lint` for linting
- `go test` with `-race` flag for concurrency testing

## Quality Gates

**Before merging any code, verify**:

1. **Constitution Compliance**:
   - [ ] Code follows SOLID principles
   - [ ] Clear, maintainable code with no unnecessary complexity
   - [ ] Appropriate test coverage following test pyramid
   - [ ] Documentation updated

2. **Technical Standards**:
   - [ ] All tests pass locally
   - [ ] Linting passes with no warnings
   - [ ] Go modules properly updated (`go mod tidy`)
   - [ ] No TODO comments without associated issues

3. **Terraform Provider Specific**:
   - [ ] Schema changes are backward compatible or properly versioned
   - [ ] Resource operations are idempotent
   - [ ] Error messages are actionable for end users
   - [ ] Examples provided for new resources/data sources

## Governance

**Authority**: This constitution supersedes all other development practices,
coding styles, or informal agreements. When conflicts arise, this document is
the source of truth.

**Amendment Process**:
1. Proposed changes MUST be documented with rationale
2. Changes MUST maintain or improve code quality and maintainability
3. All amendments MUST be based on authoritative sources (no assumptions)
4. Version number MUST be incremented per semantic versioning rules:
   - **MAJOR**: Backward incompatible principle removals/redefinitions
   - **MINOR**: New principles or materially expanded guidance
   - **PATCH**: Clarifications, wording, or non-semantic refinements

**Compliance**:
- All pull requests MUST pass constitution compliance checks
- Complexity MUST be justified in plan.md's "Complexity Tracking" section
- Team members may challenge constitution violations without repercussion
- Repeated violations require team discussion and process improvement

**Versioning**:
- Constitution version changes propagate to dependent templates automatically
- Breaking changes require migration plan for existing code
- All stakeholders MUST acknowledge major version updates

**Version**: 1.0.0 | **Ratified**: 2025-10-15 | **Last Amended**: 2025-10-15
