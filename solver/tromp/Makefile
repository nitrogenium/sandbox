# Makefile for Cuckoo Cycle lean solver

CC = gcc
CXX = g++

# Detect architecture
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M),arm64)
    # ARM/Apple Silicon
    ARCH_FLAGS = -mcpu=native
    CPPFLAGS = -DNO_SIMD
else
    # x86_64
    ARCH_FLAGS = -march=native
    CPPFLAGS = 
endif

CFLAGS = -O3 $(ARCH_FLAGS) -std=c11 -Wall -Wno-deprecated-declarations
CXXFLAGS = -O3 $(ARCH_FLAGS) -std=c++14 -Wall -Wno-deprecated-declarations -pthread $(CPPFLAGS)
INCLUDES = -I. -Icuckoo-orig/src -Icuckoo-orig/src/crypto

# Source files - always use lean solver
SOURCES = cuckoo_lean.cpp
OBJECTS = $(SOURCES:.cpp=.o)
TARGET = libcuckoo_lean.a

# Build rules
all: $(TARGET)

$(TARGET): $(OBJECTS)
	ar rcs $@ $^

%.o: %.cpp
	$(CXX) $(CXXFLAGS) $(INCLUDES) -c $< -o $@

clean:
	rm -f $(OBJECTS) $(TARGET)

# For testing
test: $(TARGET)
	$(CXX) $(CXXFLAGS) $(INCLUDES) -o test_solver test_solver.cpp -L. -lcuckoo_lean -lpthread

.PHONY: all clean test
