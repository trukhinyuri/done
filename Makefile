KOTLINC=kotlinc
KOTLINFLAGS=-include-runtime -d
JAR_NAME=out/done/done.jar
SOURCE_FILES=src/Main.kt

all: build

.PHONY: build
build:
	$(KOTLINC) $(SOURCE_FILES) $(KOTLINFLAGS) $(JAR_NAME)

.PHONY: run
run: build
	java -jar $(JAR_NAME)

.PHONY: clean
clean:
	rm -f $(JAR_NAME)
