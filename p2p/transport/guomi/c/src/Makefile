#LD_LIBRARY_PATH=$LD_LIBRARY_PATH:`pwd`/src/ ./use_mylib
all: syd
	./syd

syd: syd.cpp sydapi.h
	g++ -o syd syd.cpp libsydapi.so

clean:
	rm -rf syd
