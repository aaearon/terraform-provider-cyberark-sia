# ARK SDK API Contracts

**ARK SDK Version**: v1.5.0
**Package**: `github.com/cyberark/ark-sdk-golang/pkg/services/uap/sia/db`

## CRUD Operations

### AddPolicy
```go
func (s *ArkUAPSIADBService) AddPolicy(policy *ArkUAPSIADBAccessPolicy) (*ArkUAPSIADBAccessPolicy, error)
```
- Creates new database access policy
- Returns created policy with PolicyID populated
- Auto-sets: PolicyID, CreatedBy, UpdatedOn

### UpdatePolicy
```go
func (s *ArkUAPSIADBService) UpdatePolicy(policy *ArkUAPSIADBAccessPolicy) (*ArkUAPSIADBAccessPolicy, error)
```
- Updates existing policy
- **Constraint**: Targets map can only contain ONE workspace type per call
- Updates: UpdatedOn timestamp

### DeletePolicy
```go
func (s *ArkUAPSIADBService) DeletePolicy(req *ArkUAPDeletePolicyRequest) error
```
- Deletes policy
- Cascade deletes all principal and database assignments

### Policy (Get)
```go
func (s *ArkUAPSIADBService) Policy(req *ArkUAPGetPolicyRequest) (*ArkUAPSIADBAccessPolicy, error)
```
- Retrieves policy by ID
- Returns complete policy with all principals and targets

### ListPolicies
```go
func (s *ArkUAPSIADBService) ListPolicies() (<-chan *ArkUAPDBPolicyPage, error)
```
- Returns channel of paginated results
- Iterate with `for page := range policyPages`

## Complete Implementation Details

See research.md § ARK SDK UAP API Structure for:
- Full type hierarchy
- Field mappings
- Validation rules
- Code examples
