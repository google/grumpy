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

import types

from '__go__/grumpy' import (FunctionType, MethodType, ModuleType, StrType,  # pylint: disable=g-multiple-import
                             TracebackType, TypeType)

# Verify a sample of all types as a sanity check.
assert types.FunctionType is FunctionType
assert types.MethodType is MethodType
assert types.UnboundMethodType is MethodType
assert types.ModuleType is ModuleType
assert types.StringType is StrType
assert types.TracebackType is TracebackType
assert types.TypeType is TypeType
