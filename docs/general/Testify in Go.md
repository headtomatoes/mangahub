# Testing with Go and Testify

## Assert

if fail then continue to the end of the block

```go
func TestMultipleChecks(t *testing.T) {
    result := Calculate(5)
    assert.Equal(t, 10, result.Value)      // Test continues even if this fails
    assert.NoError(t, result.Error)        // This will still run
    assert.True(t, result.IsValid)         // And so will this
}

```

## Require

if not fulill the cond then stop

```go
func TestDependentChecks(t *testing.T) {
    user, err := GetUser(1)
    require.NoError(t, err)                // Stop here if error
    require.NotNil(t, user)                // Stop if user is nil
    assert.Equal(t, "Alice", user.Name)    // Safe to dereference now
}

```

## MOCK

mock framework for creat test double?

## SUITE

test suite support with setup/teardown hooks ?

## Best Practices

Use require for preconditions: Stop test execution early if foundational checks fail​

Use assert for multiple checks: See all failures at once when testing multiple properties​

Name tests descriptively: Use table-driven tests with clear names for each case​

Test both success and failure paths: Always test error conditions​

Use net.Pipe() for TCP testing: Avoids actual network connections in unit tests​

Mock external dependencies: Use testify/mock or interfaces for testability​

Run tests in parallel when possible: Use t.Parallel() for independent tests​

Keep tests focused: Each test should verify one behavior or scenario​

Running Tests