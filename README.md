# myhttp

## Build and Usage

You need golang  >= 1.13 in order to build the binary locally

### Local

On your terminal at this project's root folder:

  `go build`

Then (with as many URLs you like, separated by spaces):

  `./myhttp google.com apple.com`

You can also limit the number of parallel requests (default is 10):

  `./myhttp -parallel 2 google.com apple.com`

## Testing

On your terminal at this project's root folder:
  
  `go test`
  
## Assumptions 

- Retries are not needed
- In case of error, users want to see an specific message per request that can be parsed
- To keep the code simple and prevent resource exhaustion, only response bodies until 10MB will be hashed.
  Otherwise, an error will be shown
- No marshalling of any sort will be done, response bodies will be treated as raw strings
- The program does not stop on a failed request, it will try to process the remaining ones
- `-parallel` argument has to be an integer >= 1

## Design decisions, known issues and considerations

- Content-length header is used to limit response body parsing, but sometimes this header will
  not be present, which may cause high RAM usage if the body is too big.
  This is a known issue, not solved due time constraints.
- Not being able to use third party libs slows down development a lot, specially regarding testing
- Everything is in the `main` package for two reasons: Time and since it is a very small
  project, further abstractions wouldn't pay off yet. In case of more requirements come in,
  then it would be time to further split the code into modules. However, having smaller functions with specific purposes would facilitate the change
- Instead of printing as requests are finished, I chose to return the results and then print,
  to improve testability
- For the sake of simplicity, I avoided changing data structures in place. This is not a problem
  when operating with slices, but it might be for structs as they are pass by value. However, this is not a problem for the scope of this project, IMO.
- There is no current cap for max parallel requests. 
- Errors were wrapped by custom error types to prevent leaking abstractions
- Test Coverage can be improved, perhaps even a fuzzer could be added.
- Timeouts are the net/http defaults. This could be configurable in the future