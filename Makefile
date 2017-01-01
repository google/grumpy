# Copyright 2016 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# ------------------------------------------------------------------------------
# General setup
# ------------------------------------------------------------------------------

GO_ENV := $(shell go env GOOS GOARCH)
GOOS ?= $(word 1,$(GO_ENV))
GOARCH ?= $(word 2,$(GO_ENV))
ROOT_DIR := $(realpath .)
PKG_DIR := build/pkg/$(GOOS)_$(GOARCH)
PY_DIR := build/lib/python2.7/site-packages

export GOPATH := $(ROOT_DIR)/build
export PYTHONPATH := $(ROOT_DIR)/$(PY_DIR)
export PATH := $(ROOT_DIR)/build/bin:$(PATH)

COMPILER_BIN := build/bin/grumpc
COMPILER_SRCS := $(addprefix $(PY_DIR)/grumpy/compiler/,$(notdir $(shell find compiler -name '*.py' -not -name '*_test.py'))) $(PY_DIR)/grumpy/__init__.py
COMPILER_TESTS := $(patsubst %.py,grumpy/%,$(filter-out compiler/expr_visitor_test.py compiler/stmt_test.py,$(wildcard compiler/*_test.py)))
COMPILER_TEST_SRCS := $(patsubst %,$(PY_DIR)/%.py,$(COMPILER_TESTS))
COMPILER_SHARDED_TEST_SRCS := $(patsubst %,$(PY_DIR)/grumpy/compiler/%,expr_visitor_test.py stmt_test.py)
COMPILER_PASS_FILES := $(patsubst %,$(PY_DIR)/%.pass,$(COMPILER_TESTS))
COMPILER_EXPR_VISITOR_PASS_FILES := $(patsubst %,$(PY_DIR)/grumpy/compiler/expr_visitor_test.%of32.pass,$(shell seq 32))
COMPILER_STMT_PASS_FILES := $(patsubst %,$(PY_DIR)/grumpy/compiler/stmt_test.%of16.pass,$(shell seq 16))
COMPILER_D_FILES := $(patsubst %,$(PY_DIR)/%.d,$(COMPILER_TESTS))
COMPILER := $(COMPILER_BIN) $(COMPILER_SRCS)

RUNNER_BIN := build/bin/grumprun
RUNTIME_SRCS := $(addprefix build/src/grumpy/,$(notdir $(wildcard runtime/*.go)))
RUNTIME := $(PKG_DIR)/grumpy.a
RUNTIME_PASS_FILE := $(PKG_DIR)/grumpy.pass
RUNTIME_COVER_FILE := $(PKG_DIR)/grumpy.cover
RUNNER = $(RUNNER_BIN) $(COMPILER) $(RUNTIME) $(STDLIB)

GRUMPY_STDLIB_SRCS := $(shell find lib -name '*.py')
GRUMPY_STDLIB_PACKAGES := $(foreach x,$(GRUMPY_STDLIB_SRCS),$(patsubst lib/%.py,%,$(patsubst lib/%/__init__.py,%,$(x))))
THIRD_PARTY_STDLIB_SRCS := $(wildcard third_party/stdlib/*.py)
THIRD_PARTY_STDLIB_PACKAGES := $(foreach x,$(THIRD_PARTY_STDLIB_SRCS),$(patsubst third_party/stdlib/%.py,%,$(x)))
STDLIB_SRCS := $(GRUMPY_STDLIB_SRCS) $(THIRD_PARTY_STDLIB_SRCS)
STDLIB_PACKAGES := $(GRUMPY_STDLIB_PACKAGES) $(THIRD_PARTY_STDLIB_PACKAGES)
STDLIB := $(patsubst %,$(PKG_DIR)/grumpy/lib/%.a,$(STDLIB_PACKAGES))
STDLIB_TESTS := $(patsubst lib/%.py,%,$(shell find lib -name '*_test.py'))
STDLIB_PASS_FILES := $(patsubst %,$(PKG_DIR)/grumpy/lib/%.pass,$(STDLIB_TESTS))

ACCEPT_TESTS := $(patsubst %.py,%,$(wildcard testing/*.py))
ACCEPT_PASS_FILES := $(patsubst %,build/%.pass,$(ACCEPT_TESTS))

BENCHMARKS := $(patsubst %.py,%,$(wildcard benchmarks/*.py))
BENCHMARK_BINS := $(patsubst %,build/%_benchmark,$(BENCHMARKS))

TOOL_BINS = $(patsubst %,build/bin/%,benchcmp coverparse diffrange)

GOLINT_BIN = build/bin/golint
PYLINT_BIN = build/bin/pylint

all: $(COMPILER) $(RUNTIME) $(STDLIB) $(TOOL_BINS)

benchmarks: $(BENCHMARK_BINS)

clean:
	@rm -rf build

# Convenient wrapper around grumprun that avoids having to set up PATH, etc.
run: $(RUNNER)
	@$(RUNNER_BIN)

test: $(ACCEPT_PASS_FILES) $(COMPILER_PASS_FILES) $(COMPILER_EXPR_VISITOR_PASS_FILES) $(COMPILER_STMT_PASS_FILES) $(RUNTIME_PASS_FILE) $(STDLIB_PASS_FILES)

precommit: cover lint test

.PHONY: all benchmarks clean cover golint lint precommit pylint run test

# ------------------------------------------------------------------------------
# grumpc compiler
# ------------------------------------------------------------------------------

$(COMPILER_BIN) $(RUNNER_BIN) $(TOOL_BINS): build/bin/%: tools/%
	@mkdir -p build/bin
	@cp -f $< $@

$(COMPILER_SRCS) $(COMPILER_TEST_SRCS) $(COMPILER_SHARDED_TEST_SRCS): $(PY_DIR)/grumpy/%.py: %.py
	@mkdir -p $(PY_DIR)/grumpy/compiler
	@cp -f $< $@

$(COMPILER_PASS_FILES): %.pass: %.py $(COMPILER)
	@python $< -q
	@touch $@
	@echo compiler/`basename $*` PASS

$(COMPILER_D_FILES): $(PY_DIR)/%.d: $(PY_DIR)/%.py $(COMPILER_SRCS)
	@python -m modulefinder $< | awk '{if (match($$2, /^grumpy\>/)) { print "$(PY_DIR)/$*.pass: " substr($$3, length("$(ROOT_DIR)/") + 1) }}' > $@

-include $(COMPILER_D_FILES)

# Does not depend on stdlibs since it makes minimal use of them.
$(COMPILER_EXPR_VISITOR_PASS_FILES): $(PY_DIR)/grumpy/compiler/expr_visitor_test.%.pass: $(PY_DIR)/grumpy/compiler/expr_visitor_test.py $(RUNNER_BIN) $(COMPILER) $(RUNTIME)
	@python $< --shard=$*
	@touch $@
	@echo 'compiler/expr_visitor_test $* PASS'

# Does not depend on stdlibs since it makes minimal use of them.
$(COMPILER_STMT_PASS_FILES): $(PY_DIR)/grumpy/compiler/stmt_test.%.pass: $(PY_DIR)/grumpy/compiler/stmt_test.py $(RUNNER_BIN) $(COMPILER) $(RUNTIME)
	@python $< --shard=$*
	@touch $@
	@echo 'compiler/stmt_test $* PASS'

# ------------------------------------------------------------------------------
# Grumpy runtime
# ------------------------------------------------------------------------------

$(RUNTIME_SRCS): build/src/grumpy/%.go: runtime/%.go
	@mkdir -p build/src/grumpy
	@cp -f $< $@

$(RUNTIME): $(filter-out %_test.go,$(RUNTIME_SRCS))
	@mkdir -p $(PKG_DIR)
	@go tool compile -o $@ -p grumpy -complete -I $(PKG_DIR) -pack $^

$(RUNTIME_PASS_FILE): $(RUNTIME) $(filter %_test.go,$(RUNTIME_SRCS))
	@go test grumpy
	@touch $@
	@echo 'runtime/grumpy PASS'

$(RUNTIME_COVER_FILE): $(RUNTIME) $(filter %_test.go,$(RUNTIME_SRCS))
	@go test -coverprofile=$@ grumpy

cover: $(RUNTIME_COVER_FILE) $(TOOL_BINS)
	@bash -c 'comm -12 <(coverparse $< | sed "s/^grumpy/runtime/" | sort) <(git diff --dst-prefix= $(DIFF_COMMIT) | diffrange | sort)' | sort -t':' -k1,1 -k2n,2 | sed 's/$$/: missing coverage/' | tee errors.err
	@test ! -s errors.err

$(GOLINT_BIN):
	@go get -u github.com/golang/lint/golint

golint: $(GOLINT_BIN) $(PYLINT_BIN)
	@$(GOLINT_BIN) -set_exit_status runtime

$(PYLINT_BIN):
	@pip install --prefix=$(ROOT_DIR)/build pylint

pylint: $(PYLINT_BIN)
	@$(PYLINT_BIN) compiler/*.py tools/{benchcmp,coverparse,diffrange,grumpc,grumprun}

lint: golint pylint

# ------------------------------------------------------------------------------
# Standard library
# ------------------------------------------------------------------------------

$(STDLIB_PASS_FILES): $(PKG_DIR)/grumpy/lib/%.pass: $(PKG_DIR)/grumpy/lib/%.a
	@$(RUNNER_BIN) -m `echo $* | tr / .`
	@touch $@
	@echo 'lib/$* PASS'

define GRUMPY_STDLIB
build/src/grumpy/lib/$(2)/module.go: $(1) $(COMPILER)
	@mkdir -p build/src/grumpy/lib/$(2)
	@$(COMPILER_BIN) -modname=$(notdir $(2)) $(1) > $$@

build/src/grumpy/lib/$(2)/module.d: $(1)
	@mkdir -p build/src/grumpy/lib/$(2)
	@python -m modulefinder -p $(ROOT_DIR)/lib:$(ROOT_DIR)/third_party/stdlib $$< | awk '{if (($$$$1 == "m" || $$$$1 == "P") && $$$$2 != "__main__" && $$$$2 != "$(2)") {gsub(/\./, "/", $$$$2); print "$(PKG_DIR)/grumpy/lib/$(2).a: $(PKG_DIR)/grumpy/lib/" $$$$2 ".a"}}' > $$@

$(PKG_DIR)/grumpy/lib/$(2).a: build/src/grumpy/lib/$(2)/module.go $(RUNTIME)
	@mkdir -p $(PKG_DIR)/grumpy/lib/$(dir $(2))
	@go tool compile -o $$@ -p grumpy/lib/$(2) -complete -I $(PKG_DIR) -pack $$<

-include build/src/grumpy/lib/$(2)/module.d

endef

$(eval $(foreach x,$(shell seq $(words $(STDLIB_SRCS))),$(call GRUMPY_STDLIB,$(word $(x),$(STDLIB_SRCS)),$(word $(x),$(STDLIB_PACKAGES)))))

# ------------------------------------------------------------------------------
# Acceptance tests & benchmarks
# ------------------------------------------------------------------------------

$(patsubst %_test,build/%.go,$(ACCEPT_TESTS)): build/%.go: %_test.py $(COMPILER)
	@mkdir -p $(@D)
	@$(COMPILER_BIN) $< > $@

# TODO: These should not depend on stdlibs and should instead build a .d file.
$(patsubst %,build/%,$(ACCEPT_TESTS)): build/%_test: build/%.go $(RUNTIME) $(STDLIB)
	@mkdir -p $(@D)
	@go build -o $@ $<

$(ACCEPT_PASS_FILES): build/%_test.pass: build/%_test
	@$<
	@touch $@
	@echo '$*_test PASS'

$(patsubst %,build/%.go,$(BENCHMARKS)): build/%.go: %.py $(COMPILER)
	@mkdir -p $(@D)
	@$(COMPILER_BIN) $< > $@

$(BENCHMARK_BINS): build/benchmarks/%_benchmark: build/benchmarks/%.go $(RUNTIME) $(STDLIB)
	@mkdir -p $(@D)
	@go build -o $@ $<
