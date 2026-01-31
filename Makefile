build/rsmqt: main.go lib/rsmq
	 CGO_CXXFLAGS="-std=c++17 -stdlib=libc++ -fPIC -Wno-ignored-attributes -D_Bool=bool" go build -o build/rsmqt -ldflags="-s -w" .
