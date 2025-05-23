build/rsmqt:
	CGO_CXXFLAGS="-std=c++17" go build -o build/rsmqt -ldflags="-s -w" .