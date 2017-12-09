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

# Try python2 and then python if PYTHON has not been set
ifeq ($(PYTHON),)
  ifneq ($(shell which python2),)
    PYTHON = python2
  else
    PYTHON = python
  endif
endif
PYTHON_BIN := $(shell which $(PYTHON))
PYTHON_VER := $(word 2,$(shell $(PYTHON) -V 2>&1))
GO_REQ_MAJ := 1
GO_REQ_MIN := 9
GO_MAJ_MIN := $(subst go,, $(word 3,$(shell go version 2>&1)) )
GO_MAJ := $(word 1,$(subst ., ,$(GO_MAJ_MIN) ))
GO_MIN := $(word 2,$(subst ., ,$(GO_MAJ_MIN) ))

ifeq ($(filter 2.7.%,$(PYTHON_VER)),)
  $(error unsupported Python version $(PYTHON_VER), Grumpy only supports 2.7.x. To use a different python binary such as python2, run: 'make PYTHON=python2 ...')
endif

ifneq ($(shell test $(GO_MAJ) -ge $(GO_REQ_MAJ) -a $(GO_MIN) -ge $(GO_REQ_MIN) && echo ok),ok)
  $(error unsupported Go version $(GO_VER), Grumpy requires at least $(GO_REQ_MAJ).$(GO_REQ_MIN). Please update Go)
endif

PY_DIR := build/lib/python2.7/site-packages
PY_INSTALL_DIR := $(shell $(PYTHON) -c "from distutils.sysconfig import get_python_lib; print(get_python_lib())")

export GOPATH := $(ROOT_DIR)/build
export PYTHONPATH := $(ROOT_DIR)/$(PY_DIR)
export PATH := $(ROOT_DIR)/build/bin:$(PATH)

GOPATH_PY_ROOT := $(GOPATH)/src/__python__

PYTHONPARSER_SRCS := $(patsubst third_party/%,$(PY_DIR)/grumpy/%,$(wildcard third_party/pythonparser/*.py))

COMPILER_BIN := build/bin/grumpc
COMPILER_SRCS := $(addprefix $(PY_DIR)/grumpy/compiler/,$(notdir $(shell find compiler -name '*.py' -not -name '*_test.py'))) $(PY_DIR)/grumpy/__init__.py
COMPILER_TESTS := $(patsubst %.py,grumpy/%,$(filter-out compiler/expr_visitor_test.py compiler/stmt_test.py,$(wildcard compiler/*_test.py)))
COMPILER_TEST_SRCS := $(patsubst %,$(PY_DIR)/%.py,$(COMPILER_TESTS))
COMPILER_SHARDED_TEST_SRCS := $(patsubst %,$(PY_DIR)/grumpy/compiler/%,expr_visitor_test.py stmt_test.py)
COMPILER_PASS_FILES := $(patsubst %,$(PY_DIR)/%.pass,$(COMPILER_TESTS))
COMPILER_EXPR_VISITOR_PASS_FILES := $(patsubst %,$(PY_DIR)/grumpy/compiler/expr_visitor_test.%of32.pass,$(shell seq 32))
COMPILER_STMT_PASS_FILES := $(patsubst %,$(PY_DIR)/grumpy/compiler/stmt_test.%of16.pass,$(shell seq 16))
COMPILER_D_FILES := $(patsubst %,$(PY_DIR)/%.d,$(COMPILER_TESTS))
COMPILER := $(COMPILER_BIN) $(COMPILER_SRCS) $(PYTHONPARSER_SRCS)

PKGC_BIN := build/bin/pkgc

RUNNER_BIN := build/bin/grumprun
RUNTIME_SRCS := $(addprefix build/src/grumpy/,$(notdir $(wildcard runtime/*.go)))
RUNTIME := $(PKG_DIR)/grumpy.a
RUNTIME_PASS_FILE := build/runtime.pass
RUNTIME_COVER_FILE := $(PKG_DIR)/grumpy.cover
RUNNER = $(RUNNER_BIN) $(COMPILER) $(RUNTIME) $(STDLIB)

LIB_SRCS := $(patsubst lib/%,$(GOPATH_PY_ROOT)/%,$(shell find lib -name '*.py'))
THIRD_PARTY_STDLIB_SRCS := $(patsubst third_party/stdlib/%,$(GOPATH_PY_ROOT)/%,$(shell find third_party/stdlib -name '*.py'))
THIRD_PARTY_PYPY_SRCS := $(patsubst third_party/pypy/%,$(GOPATH_PY_ROOT)/%,$(shell find third_party/pypy -name '*.py'))
THIRD_PARTY_OUROBOROS_SRCS := $(patsubst third_party/ouroboros/%,$(GOPATH_PY_ROOT)/%,$(shell find third_party/ouroboros -name '*.py'))
STDLIB_SRCS := $(LIB_SRCS) $(THIRD_PARTY_STDLIB_SRCS) $(THIRD_PARTY_PYPY_SRCS) $(THIRD_PARTY_OUROBOROS_SRCS)

STDLIB_PACKAGES := $(patsubst $(GOPATH_PY_ROOT)/%.py,%,$(patsubst $(GOPATH_PY_ROOT)/%/__init__.py,%,$(STDLIB_SRCS)))
STDLIB := $(patsubst %,$(PKG_DIR)/__python__/%.a,$(STDLIB_PACKAGES))
STDLIB_TESTS := \
  itertools_test \
  math_test \
  os/path_test \
  os_test \
  random_test \
  re_tests \
  sys_test \
  tempfile_test \
  test/test_bisect \
  test/test_colorsys \
  test/test_datetime \
  test/test_dict \
  test/test_dircache \
  test/test_dummy_thread \
  test/test_fpformat \
  test/test_genericpath \
  test/test_list \
  test/test_md5 \
  test/test_mimetools \
  test/test_mutex \
  test/test_operator \
  test/test_quopri \
  test/test_queue \
  test/test_rfc822 \
  test/test_sched \
  test/test_select \
  test/test_slice \
  test/test_stat \
  test/test_string \
  test/test_threading \
  test/test_tuple \
  test/test_uu \
  time_test \
  types_test \
  weetest_test
STDLIB_PASS_FILES := $(patsubst %,build/testing/%.pass,$(notdir $(STDLIB_TESTS)))

ACCEPT_TESTS := $(patsubst %.py,%,$(wildcard testing/*.py))
ACCEPT_PASS_FILES := $(patsubst %,build/%.pass,$(ACCEPT_TESTS))
ACCEPT_PY_PASS_FILES := $(patsubst %,build/%_py.pass,$(filter-out %/native_test,$(ACCEPT_TESTS)))

BENCHMARKS := $(patsubst %.py,%,$(wildcard benchmarks/*.py))
BENCHMARK_BINS := $(patsubst %,build/%_benchmark,$(BENCHMARKS))

TOOL_BINS = $(patsubst %,build/bin/%,benchcmp coverparse diffrange genmake pydeps)

GOLINT_BIN = build/bin/golint
PYLINT_BIN = build/bin/pylint

all: $(COMPILER) $(RUNNER) $(RUNTIME) $(TOOL_BINS)

benchmarks: $(BENCHMARK_BINS)

clean:
	@rm -rf build

# Convenient wrapper around grumprun that avoids having to set up PATH, etc.
run: $(RUNNER)
	@$(RUNNER_BIN)

test: $(ACCEPT_PASS_FILES) $(ACCEPT_PY_PASS_FILES) $(COMPILER_PASS_FILES) $(COMPILER_EXPR_VISITOR_PASS_FILES) $(COMPILER_STMT_PASS_FILES) $(RUNTIME_PASS_FILE) $(STDLIB_PASS_FILES)

precommit: cover gofmt lint test

.PHONY: all benchmarks clean cover gofmt golint install lint precommit pylint run test

# ------------------------------------------------------------------------------
# grumpc compiler
# ------------------------------------------------------------------------------

$(COMPILER_BIN) $(RUNNER_BIN) $(TOOL_BINS): build/bin/%: tools/%
	@mkdir -p build/bin
	@cp -f $< $@
	@sed -i.bak -e '1s@/usr/bin/env python@$(PYTHON_BIN)@' $@
	@rm -f $@.bak

$(COMPILER_SRCS) $(COMPILER_TEST_SRCS) $(COMPILER_SHARDED_TEST_SRCS): $(PY_DIR)/grumpy/%.py: %.py
	@mkdir -p $(PY_DIR)/grumpy/compiler
	@cp -f $< $@

$(COMPILER_PASS_FILES): %.pass: %.py $(COMPILER) $(COMPILER_TEST_SRCS)
	@$(PYTHON) $< -q
	@touch $@
	@echo compiler/`basename $*` PASS

# NOTE: In the regex below we use (\.|$) instead of \> because the latter is
# not available in the awk available on OS X.
$(COMPILER_D_FILES): $(PY_DIR)/%.d: $(PY_DIR)/%.py $(COMPILER_SRCS) $(PYTHONPARSER_SRCS)
	@$(PYTHON) -m modulefinder $< | awk '{if (match($$2, /^grumpy(\.|$$)/)) { print "$(PY_DIR)/$*.pass: " substr($$3, length("$(ROOT_DIR)/") + 1) }}' > $@

-include $(COMPILER_D_FILES)

# Does not depend on stdlibs since it makes minimal use of them.
$(COMPILER_EXPR_VISITOR_PASS_FILES): $(PY_DIR)/grumpy/compiler/expr_visitor_test.%.pass: $(PY_DIR)/grumpy/compiler/expr_visitor_test.py $(RUNNER_BIN) $(COMPILER) $(RUNTIME) $(PKG_DIR)/__python__/traceback.a
	@$(PYTHON) $< --shard=$*
	@touch $@
	@echo 'compiler/expr_visitor_test $* PASS'

COMPILER_STMT_PASS_FILE_DEPS := \
  $(PKG_DIR)/__python__/__go__/grumpy.a \
  $(PKG_DIR)/__python__/__go__/os.a \
  $(PKG_DIR)/__python__/__go__/runtime.a \
  $(PKG_DIR)/__python__/__go__/time.a \
  $(PKG_DIR)/__python__/__go__/unicode.a \
  $(PKG_DIR)/__python__/sys.a \
  $(PKG_DIR)/__python__/traceback.a

# Does not depend on stdlibs since it makes minimal use of them.
$(COMPILER_STMT_PASS_FILES): $(PY_DIR)/grumpy/compiler/stmt_test.%.pass: $(PY_DIR)/grumpy/compiler/stmt_test.py $(RUNNER_BIN) $(COMPILER) $(RUNTIME) $(COMPILER_STMT_PASS_FILE_DEPS)
	@$(PYTHON) $< --shard=$*
	@touch $@
	@echo 'compiler/stmt_test $* PASS'

$(PKGC_BIN): tools/pkgc.go
	@mkdir -p $(@D)
	@go build -o $@ $<

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

build/gofmt.diff: $(wildcard runtime/*.go)
	@gofmt -d $^ > $@

gofmt: build/gofmt.diff
	@if [ -s $< ]; then echo 'gofmt found errors, run: gofmt -w $(ROOT_DIR)/runtime/*.go'; false; fi

$(GOLINT_BIN):
	@go get -u github.com/golang/lint/golint

golint: $(GOLINT_BIN)
	@$(GOLINT_BIN) -set_exit_status runtime

# Don't use system pip for this since behavior varies a lot between versions.
# Instead build pylint from source. Dependencies will be fetched by distutils.
$(PYLINT_BIN):
	@mkdir -p build/third_party
	@cd build/third_party && curl -sL https://pypi.io/packages/source/p/pylint/pylint-1.6.4.tar.gz | tar -zx
	@cd build/third_party/pylint-1.6.4 && $(PYTHON) setup.py install --prefix $(ROOT_DIR)/build

pylint: $(PYLINT_BIN) $(COMPILER_SRCS) $(PYTHONPARSER_SRCS) $(COMPILER_BIN) $(RUNNER_BIN) $(TOOL_BINS)
	@$(PYTHON) $(PYLINT_BIN) $(COMPILER_SRCS) $(COMPILER_BIN) $(RUNNER_BIN) $(TOOL_BINS)

lint: golint pylint

# ------------------------------------------------------------------------------
# Native modules
# ------------------------------------------------------------------------------

$(PKG_DIR)/__python__/__go__/%.a: build/src/__python__/__go__/%/module.go $(RUNTIME)
	@mkdir -p $(@D)
	@go install __python__/__go__/$*

build/src/__python__/__go__/%/module.go: $(PKGC_BIN) $(RUNTIME)
	@mkdir -p $(@D)
	@$(PKGC_BIN) $* > $@

$(PKG_DIR)/__python__/__go__/grumpy.a: $(RUNTIME)

.PRECIOUS: build/src/__python__/__go__/%/module.go $(PKG_DIR)/__python__/__go__/%.a

# ------------------------------------------------------------------------------
# Standard library
# ------------------------------------------------------------------------------

$(LIB_SRCS): $(GOPATH_PY_ROOT)/%: lib/%
	@mkdir -p $(@D)
	@cp -f $< $@

$(THIRD_PARTY_STDLIB_SRCS): $(GOPATH_PY_ROOT)/%: third_party/stdlib/%
	@mkdir -p $(@D)
	@cp -f $< $@

$(THIRD_PARTY_PYPY_SRCS): $(GOPATH_PY_ROOT)/%: third_party/pypy/%
	@mkdir -p $(@D)
	@cp -f $< $@


$(THIRD_PARTY_OUROBOROS_SRCS): $(GOPATH_PY_ROOT)/%: third_party/ouroboros/%
	@mkdir -p $(@D)
	@cp -f $< $@

build/stdlib.mk: build/bin/genmake | $(STDLIB_SRCS)
	@genmake build > $@

-include build/stdlib.mk

$(patsubst %,build/src/__python__/%/module.go,$(STDLIB_PACKAGES)): $(COMPILER)
$(patsubst %,build/src/__python__/%/module.d,$(STDLIB_PACKAGES)): build/bin/pydeps $(PYTHONPARSER_SRCS) $(COMPILER)
$(patsubst %,$(PKG_DIR)/__python__/%.a,$(STDLIB_PACKAGES)): $(RUNTIME)

define GRUMPY_STDLIB_TEST
build/testing/$(notdir $(1)).pass: $(RUNTIME) $(PKG_DIR)/__python__/$(1).a $(RUNNER_BIN) $(PKG_DIR)/__python__/traceback.a
	@mkdir -p $$(@D)
	@$(RUNNER_BIN) -m $(subst /,.,$(1))
	@touch $$@
	@echo 'lib/$(1) PASS'

endef

$(eval $(foreach x,$(STDLIB_TESTS),$(call GRUMPY_STDLIB_TEST,$(x))))

# ------------------------------------------------------------------------------
# Acceptance tests & benchmarks
# ------------------------------------------------------------------------------

$(PY_DIR)/weetest.py: lib/weetest.py
	@cp -f $< $@

$(PYTHONPARSER_SRCS): $(PY_DIR)/grumpy/%: third_party/%
	@mkdir -p $(@D)
	@cp -f $< $@

$(ACCEPT_PASS_FILES): build/%_test.pass: %_test.py $(RUNTIME) $(STDLIB) $(RUNNER_BIN)
	@mkdir -p $(@D)
	@$(RUNNER_BIN) < $<
	@touch $@
	@echo '$*_test PASS'

NATIVE_TEST_DEPS := \
  $(PKG_DIR)/__python__/__go__/encoding/csv.a \
  $(PKG_DIR)/__python__/__go__/image.a \
  $(PKG_DIR)/__python__/__go__/math.a \
  $(PKG_DIR)/__python__/__go__/strings.a

build/testing/native_test.pass: $(NATIVE_TEST_DEPS)

$(ACCEPT_PY_PASS_FILES): build/%_py.pass: %.py $(PY_DIR)/weetest.py
	@mkdir -p $(@D)
	@$(PYTHON) $<
	@touch $@
	@echo '$*_py PASS'

$(patsubst %,build/%.go,$(BENCHMARKS)): build/%.go: %.py $(COMPILER)
	@mkdir -p $(@D)
	@$(COMPILER_BIN) $< > $@

$(BENCHMARK_BINS): build/benchmarks/%_benchmark: build/benchmarks/%.go $(RUNTIME) $(STDLIB)
	@mkdir -p $(@D)
	@go build -o $@ $<

# ------------------------------------------------------------------------------
# Installation
# ------------------------------------------------------------------------------

install: $(RUNNER_BIN) $(COMPILER) $(RUNTIME) $(STDLIB)
	# Binary executables
	install -d "$(DESTDIR)/usr/bin"
	install -m755 build/bin/grumpc "$(DESTDIR)/usr/bin/grumpc"
	install -m755 build/bin/grumprun "$(DESTDIR)/usr/bin/grumprun"
	# Python module
	install -d "$(DESTDIR)"{/usr/lib/go,"$(PY_INSTALL_DIR)"}
	cp -rv "$(PY_DIR)/grumpy" "$(DESTDIR)$(PY_INSTALL_DIR)"
	# Go package and sources
	install -d "$(DESTDIR)/usr/lib/go/"
	cp -rv build/pkg build/src "$(DESTDIR)/usr/lib/go/"
