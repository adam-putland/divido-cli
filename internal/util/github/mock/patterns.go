package mock

var GetReposReleasesLatestByOwnerByRepo = EndpointPattern{
	Pattern: "/repos/{owner}/{repo}/releases/latest",
	Method:  "GET",
}
