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

import sys

print sys.maxint

failed_modules = []

try:
    import __builtin__  # noqa
    print('__builtin__ imported')
except ImportError:
    failed_modules.append('__builtin__')

try:
    import __future__  # noqa
    print('__future__ imported')
except ImportError:
    failed_modules.append('__future__')

try:
    import __main__  # noqa
    print('__main__ imported')
except ImportError:
    failed_modules.append('__main__')

try:
    import _winreg  # noqa
    print('_winreg imported')
except ImportError:
    failed_modules.append('_winreg')

try:
    import abc  # noqa
    print('abc imported')
except ImportError:
    failed_modules.append('abc')

try:
    import aepack  # noqa
    print('aepack imported')
except ImportError:
    failed_modules.append('aepack')

try:
    import aetools  # noqa
    print('aetools imported')
except ImportError:
    failed_modules.append('aetools')

try:
    import aetypes  # noqa
    print('aetypes imported')
except ImportError:
    failed_modules.append('aetypes')

try:
    import aifc  # noqa
    print('aifc imported')
except ImportError:
    failed_modules.append('aifc')

try:
    import AL  # noqa
    print('AL imported')
except ImportError:
    failed_modules.append('AL')

try:
    import al  # noqa
    print('al imported')
except ImportError:
    failed_modules.append('al')

try:
    import anydbm  # noqa
    print('anydbm imported')
except ImportError:
    failed_modules.append('anydbm')

try:
    import applesingle  # noqa
    print('applesingle imported')
except ImportError:
    failed_modules.append('applesingle')

try:
    import argparse  # noqa
    print('argparse imported')
except ImportError:
    failed_modules.append('argparse')

try:
    import array  # noqa
    print('array imported')
except ImportError:
    failed_modules.append('array')

try:
    import ast  # noqa
    print('ast imported')
except ImportError:
    failed_modules.append('ast')

try:
    import asynchat  # noqa
    print('asynchat imported')
except ImportError:
    failed_modules.append('asynchat')

try:
    import asyncore  # noqa
    print('asyncore imported')
except ImportError:
    failed_modules.append('asyncore')

try:
    import atexit  # noqa
    print('atexit imported')
except ImportError:
    failed_modules.append('atexit')

try:
    import audioop  # noqa
    print('audioop imported')
except ImportError:
    failed_modules.append('audioop')

try:
    import autoGIL  # noqa
    print('autoGIL imported')
except ImportError:
    failed_modules.append('autoGIL')

try:
    import base64  # noqa
    print('base64 imported')
except ImportError:
    failed_modules.append('base64')

try:
    import BaseHTTPServer  # noqa
    print('BaseHTTPServer imported')
except ImportError:
    failed_modules.append('BaseHTTPServer')

try:
    import Bastion  # noqa
    print('Bastion imported')
except ImportError:
    failed_modules.append('Bastion')

try:
    import bdb  # noqa
    print('bdb imported')
except ImportError:
    failed_modules.append('bdb')

try:
    import binascii  # noqa
    print('binascii imported')
except ImportError:
    failed_modules.append('binascii')

try:
    import binhex  # noqa
    print('binhex imported')
except ImportError:
    failed_modules.append('binhex')

try:
    import bisect  # noqa
    print('bisect imported')
except ImportError:
    failed_modules.append('bisect')

try:
    import bsddb  # noqa
    print('bsddb imported')
except ImportError:
    failed_modules.append('bsddb')

try:
    import buildtools  # noqa
    print('buildtools imported')
except ImportError:
    failed_modules.append('buildtools')

try:
    import bz2  # noqa
    print('bz2 imported')
except ImportError:
    failed_modules.append('bz2')

try:
    import calendar  # noqa
    print('calendar imported')
except ImportError:
    failed_modules.append('calendar')

try:
    import Carbon  # noqa
    print('Carbon imported')
except ImportError:
    failed_modules.append('Carbon')

try:
    import cd  # noqa
    print('cd imported')
except ImportError:
    failed_modules.append('cd')

try:
    import cfmfile  # noqa
    print('cfmfile imported')
except ImportError:
    failed_modules.append('cfmfile')

try:
    import cgi  # noqa
    print('cgi imported')
except ImportError:
    failed_modules.append('cgi')

try:
    import CGIHTTPServer  # noqa
    print('CGIHTTPServer imported')
except ImportError:
    failed_modules.append('CGIHTTPServer')

try:
    import cgitb  # noqa
    print('cgitb imported')
except ImportError:
    failed_modules.append('cgitb')

try:
    import chunk  # noqa
    print('chunk imported')
except ImportError:
    failed_modules.append('chunk')

try:
    import cmath  # noqa
    print('cmath imported')
except ImportError:
    failed_modules.append('cmath')

try:
    import cmd  # noqa
    print('cmd imported')
except ImportError:
    failed_modules.append('cmd')

try:
    import code  # noqa
    print('code imported')
except ImportError:
    failed_modules.append('code')

try:
    import codecs  # noqa
    print('codecs imported')
except ImportError:
    failed_modules.append('codecs')

try:
    import codeop  # noqa
    print('codeop imported')
except ImportError:
    failed_modules.append('codeop')

try:
    import collections  # noqa
    print('collections imported')
except ImportError:
    failed_modules.append('collections')

try:
    import ColorPicker  # noqa
    print('ColorPicker imported')
except ImportError:
    failed_modules.append('ColorPicker')

try:
    import colorsys  # noqa
    print('colorsys imported')
except ImportError:
    failed_modules.append('colorsys')

try:
    import commands  # noqa
    print('commands imported')
except ImportError:
    failed_modules.append('commands')

try:
    import compileall  # noqa
    print('compileall imported')
except ImportError:
    failed_modules.append('compileall')

try:
    import compiler  # noqa
    print('compiler imported')
except ImportError:
    failed_modules.append('compiler')

try:
    import ConfigParser  # noqa
    print('ConfigParser imported')
except ImportError:
    failed_modules.append('ConfigParser')

try:
    import contextlib  # noqa
    print('contextlib imported')
except ImportError:
    failed_modules.append('contextlib')

try:
    import Cookie  # noqa
    print('Cookie imported')
except ImportError:
    failed_modules.append('Cookie')

try:
    import cookielib  # noqa
    print('cookielib imported')
except ImportError:
    failed_modules.append('cookielib')

try:
    import copy  # noqa
    print('copy imported')
except ImportError:
    failed_modules.append('copy')

try:
    import copy_reg  # noqa
    print('copy_reg imported')
except ImportError:
    failed_modules.append('copy_reg')

try:
    import cPickle  # noqa
    print('cPickle imported')
except ImportError:
    failed_modules.append('cPickle')

try:
    import cProfile  # noqa
    print('cProfile imported')
except ImportError:
    failed_modules.append('cProfile')

try:
    import crypt  # noqa
    print('crypt imported')
except ImportError:
    failed_modules.append('crypt')

try:
    import cStringIO  # noqa
    print('cStringIO imported')
except ImportError:
    failed_modules.append('cStringIO')

try:
    import csv  # noqa
    print('csv imported')
except ImportError:
    failed_modules.append('csv')

try:
    import ctypes  # noqa
    print('ctypes imported')
except ImportError:
    failed_modules.append('ctypes')

try:
    import curses  # noqa
    print('curses imported')
except ImportError:
    failed_modules.append('curses')

try:
    import datetime  # noqa
    print('datetime imported')
except ImportError:
    failed_modules.append('datetime')

try:
    import dbhash  # noqa
    print('dbhash imported')
except ImportError:
    failed_modules.append('dbhash')

try:
    import dbm  # noqa
    print('dbm imported')
except ImportError:
    failed_modules.append('dbm')

try:
    import decimal  # noqa
    print('decimal imported')
except ImportError:
    failed_modules.append('decimal')

try:
    import DEVICE  # noqa
    print('DEVICE imported')
except ImportError:
    failed_modules.append('DEVICE')

try:
    import difflib  # noqa
    print('difflib imported')
except ImportError:
    failed_modules.append('difflib')

try:
    import dircache  # noqa
    print('dircache imported')
except ImportError:
    failed_modules.append('dircache')

try:
    import dis  # noqa
    print('dis imported')
except ImportError:
    failed_modules.append('dis')

try:
    import distutils  # noqa
    print('distutils imported')
except ImportError:
    failed_modules.append('distutils')

try:
    import dl  # noqa
    print('dl imported')
except ImportError:
    failed_modules.append('dl')

try:
    import doctest  # noqa
    print('doctest imported')
except ImportError:
    failed_modules.append('doctest')

try:
    import DocXMLRPCServer  # noqa
    print('DocXMLRPCServer imported')
except ImportError:
    failed_modules.append('DocXMLRPCServer')

try:
    import dumbdbm  # noqa
    print('dumbdbm imported')
except ImportError:
    failed_modules.append('dumbdbm')

try:
    import dummy_thread  # noqa
    print('dummy_thread imported')
except ImportError:
    failed_modules.append('dummy_thread')

try:
    import dummy_threading  # noqa
    print('dummy_threading imported')
except ImportError:
    failed_modules.append('dummy_threading')

try:
    import EasyDialogs  # noqa
    print('EasyDialogs imported')
except ImportError:
    failed_modules.append('EasyDialogs')

try:
    import email  # noqa
    print('email imported')
except ImportError:
    failed_modules.append('email')

try:
    import encodings  # noqa
    print('encodings imported')
except ImportError:
    failed_modules.append('encodings')

try:
    import ensurepip  # noqa
    print('ensurepip imported')
except ImportError:
    failed_modules.append('ensurepip')

try:
    import errno  # noqa
    print('errno imported')
except ImportError:
    failed_modules.append('errno')

try:
    import exceptions  # noqa
    print('exceptions imported')
except ImportError:
    failed_modules.append('exceptions')

try:
    import fcntl  # noqa
    print('fcntl imported')
except ImportError:
    failed_modules.append('fcntl')

try:
    import filecmp  # noqa
    print('filecmp imported')
except ImportError:
    failed_modules.append('filecmp')

try:
    import fileinput  # noqa
    print('fileinput imported')
except ImportError:
    failed_modules.append('fileinput')

try:
    import findertools  # noqa
    print('findertools imported')
except ImportError:
    failed_modules.append('findertools')

try:
    import FL  # noqa
    print('FL imported')
except ImportError:
    failed_modules.append('FL')

try:
    import fl  # noqa
    print('fl imported')
except ImportError:
    failed_modules.append('fl')

try:
    import flp  # noqa
    print('flp imported')
except ImportError:
    failed_modules.append('flp')

try:
    import fm  # noqa
    print('fm imported')
except ImportError:
    failed_modules.append('fm')

try:
    import fnmatch  # noqa
    print('fnmatch imported')
except ImportError:
    failed_modules.append('fnmatch')

try:
    import formatter  # noqa
    print('formatter imported')
except ImportError:
    failed_modules.append('formatter')

try:
    import fpectl  # noqa
    print('fpectl imported')
except ImportError:
    failed_modules.append('fpectl')

try:
    import fpformat  # noqa
    print('fpformat imported')
except ImportError:
    failed_modules.append('fpformat')

try:
    import fractions  # noqa
    print('fractions imported')
except ImportError:
    failed_modules.append('fractions')

try:
    import FrameWork  # noqa
    print('FrameWork imported')
except ImportError:
    failed_modules.append('FrameWork')

try:
    import ftplib  # noqa
    print('ftplib imported')
except ImportError:
    failed_modules.append('ftplib')

try:
    import functools  # noqa
    print('functools imported')
except ImportError:
    failed_modules.append('functools')

try:
    import future_builtins  # noqa
    print('future_builtins imported')
except ImportError:
    failed_modules.append('future_builtins')

try:
    import gc  # noqa
    print('gc imported')
except ImportError:
    failed_modules.append('gc')

try:
    import gdbm  # noqa
    print('gdbm imported')
except ImportError:
    failed_modules.append('gdbm')

try:
    import gensuitemodule  # noqa
    print('gensuitemodule imported')
except ImportError:
    failed_modules.append('gensuitemodule')

try:
    import getopt  # noqa
    print('getopt imported')
except ImportError:
    failed_modules.append('getopt')

try:
    import getpass  # noqa
    print('getpass imported')
except ImportError:
    failed_modules.append('getpass')

try:
    import gettext  # noqa
    print('gettext imported')
except ImportError:
    failed_modules.append('gettext')

try:
    import gl  # noqa
    print('gl imported')
except ImportError:
    failed_modules.append('gl')

try:
    import GL  # noqa
    print('GL imported')
except ImportError:
    failed_modules.append('GL')

try:
    import glob  # noqa
    print('glob imported')
except ImportError:
    failed_modules.append('glob')

try:
    import grp  # noqa
    print('grp imported')
except ImportError:
    failed_modules.append('grp')

try:
    import gzip  # noqa
    print('gzip imported')
except ImportError:
    failed_modules.append('gzip')

try:
    import hashlib  # noqa
    print('hashlib imported')
except ImportError:
    failed_modules.append('hashlib')

try:
    import heapq  # noqa
    print('heapq imported')
except ImportError:
    failed_modules.append('heapq')

try:
    import hmac  # noqa
    print('hmac imported')
except ImportError:
    failed_modules.append('hmac')

try:
    import hotshot  # noqa
    print('hotshot imported')
except ImportError:
    failed_modules.append('hotshot')

try:
    import htmlentitydefs  # noqa
    print('htmlentitydefs imported')
except ImportError:
    failed_modules.append('htmlentitydefs')

try:
    import htmllib  # noqa
    print('htmllib imported')
except ImportError:
    failed_modules.append('htmllib')

try:
    import HTMLParser  # noqa
    print('HTMLParser imported')
except ImportError:
    failed_modules.append('HTMLParser')

try:
    import httplib  # noqa
    print('httplib imported')
except ImportError:
    failed_modules.append('httplib')

try:
    import ic  # noqa
    print('ic imported')
except ImportError:
    failed_modules.append('ic')

try:
    import icopen  # noqa
    print('icopen imported')
except ImportError:
    failed_modules.append('icopen')

try:
    import imageop  # noqa
    print('imageop imported')
except ImportError:
    failed_modules.append('imageop')

try:
    import imaplib  # noqa
    print('imaplib imported')
except ImportError:
    failed_modules.append('imaplib')

try:
    import imgfile  # noqa
    print('imgfile imported')
except ImportError:
    failed_modules.append('imgfile')

try:
    import imghdr  # noqa
    print('imghdr imported')
except ImportError:
    failed_modules.append('imghdr')

try:
    import imp  # noqa
    print('imp imported')
except ImportError:
    failed_modules.append('imp')

try:
    import importlib  # noqa
    print('importlib imported')
except ImportError:
    failed_modules.append('importlib')

try:
    import imputil  # noqa
    print('imputil imported')
except ImportError:
    failed_modules.append('imputil')

try:
    import inspect  # noqa
    print('inspect imported')
except ImportError:
    failed_modules.append('inspect')

try:
    import io  # noqa
    print('io imported')
except ImportError:
    failed_modules.append('io')

try:
    import itertools  # noqa
    print('itertools imported')
except ImportError:
    failed_modules.append('itertools')

try:
    import jpeg  # noqa
    print('jpeg imported')
except ImportError:
    failed_modules.append('jpeg')

try:
    import json  # noqa
    print('json imported')
except ImportError:
    failed_modules.append('json')

try:
    import keyword  # noqa
    print('keyword imported')
except ImportError:
    failed_modules.append('keyword')

try:
    import lib2to3  # noqa
    print('lib2to3 imported')
except ImportError:
    failed_modules.append('lib2to3')

try:
    import linecache  # noqa
    print('linecache imported')
except ImportError:
    failed_modules.append('linecache')

try:
    import locale  # noqa
    print('locale imported')
except ImportError:
    failed_modules.append('locale')

try:
    import logging  # noqa
    print('logging imported')
except ImportError:
    failed_modules.append('logging')

try:
    import macerrors  # noqa
    print('macerrors imported')
except ImportError:
    failed_modules.append('macerrors')

try:
    import MacOS  # noqa
    print('MacOS imported')
except ImportError:
    failed_modules.append('MacOS')

try:
    import macostools  # noqa
    print('macostools imported')
except ImportError:
    failed_modules.append('macostools')

try:
    import macpath  # noqa
    print('macpath imported')
except ImportError:
    failed_modules.append('macpath')

try:
    import macresource  # noqa
    print('macresource imported')
except ImportError:
    failed_modules.append('macresource')

try:
    import mailbox  # noqa
    print('mailbox imported')
except ImportError:
    failed_modules.append('mailbox')

try:
    import mailcap  # noqa
    print('mailcap imported')
except ImportError:
    failed_modules.append('mailcap')

try:
    import marshal  # noqa
    print('marshal imported')
except ImportError:
    failed_modules.append('marshal')

try:
    import math  # noqa
    print('math imported')
except ImportError:
    failed_modules.append('math')

try:
    import md5  # noqa
    print('md5 imported')
except ImportError:
    failed_modules.append('md5')

try:
    import mhlib  # noqa
    print('mhlib imported')
except ImportError:
    failed_modules.append('mhlib')

try:
    import mimetools  # noqa
    print('mimetools imported')
except ImportError:
    failed_modules.append('mimetools')

try:
    import mimetypes  # noqa
    print('mimetypes imported')
except ImportError:
    failed_modules.append('mimetypes')

try:
    import MimeWriter  # noqa
    print('MimeWriter imported')
except ImportError:
    failed_modules.append('MimeWriter')

try:
    import mimify  # noqa
    print('mimify imported')
except ImportError:
    failed_modules.append('mimify')

try:
    import MiniAEFrame  # noqa
    print('MiniAEFrame imported')
except ImportError:
    failed_modules.append('MiniAEFrame')

try:
    import mmap  # noqa
    print('mmap imported')
except ImportError:
    failed_modules.append('mmap')

try:
    import modulefinder  # noqa
    print('modulefinder imported')
except ImportError:
    failed_modules.append('modulefinder')

try:
    import msilib  # noqa
    print('msilib imported')
except ImportError:
    failed_modules.append('msilib')

try:
    import msvcrt  # noqa
    print('msvcrt imported')
except ImportError:
    failed_modules.append('msvcrt')

try:
    import multifile  # noqa
    print('multifile imported')
except ImportError:
    failed_modules.append('multifile')

try:
    import multiprocessing  # noqa
    print('multiprocessing imported')
except ImportError:
    failed_modules.append('multiprocessing')

try:
    import mutex  # noqa
    print('mutex imported')
except ImportError:
    failed_modules.append('mutex')

try:
    import Nav  # noqa
    print('Nav imported')
except ImportError:
    failed_modules.append('Nav')

try:
    import netrc  # noqa
    print('netrc imported')
except ImportError:
    failed_modules.append('netrc')

try:
    import new  # noqa
    print('new imported')
except ImportError:
    failed_modules.append('new')

try:
    import nis  # noqa
    print('nis imported')
except ImportError:
    failed_modules.append('nis')

try:
    import nntplib  # noqa
    print('nntplib imported')
except ImportError:
    failed_modules.append('nntplib')

try:
    import numbers  # noqa
    print('numbers imported')
except ImportError:
    failed_modules.append('numbers')

try:
    import operator  # noqa
    print('operator imported')
except ImportError:
    failed_modules.append('operator')

try:
    import optparse  # noqa
    print('optparse imported')
except ImportError:
    failed_modules.append('optparse')

try:
    import os  # noqa
    print('os imported')
except ImportError:
    failed_modules.append('os')

try:
    import ossaudiodev  # noqa
    print('ossaudiodev imported')
except ImportError:
    failed_modules.append('ossaudiodev')

try:
    import parser  # noqa
    print('parser imported')
except ImportError:
    failed_modules.append('parser')

try:
    import pdb  # noqa
    print('pdb imported')
except ImportError:
    failed_modules.append('pdb')

try:
    import pickle  # noqa
    print('pickle imported')
except ImportError:
    failed_modules.append('pickle')

try:
    import pickletools  # noqa
    print('pickletools imported')
except ImportError:
    failed_modules.append('pickletools')

try:
    import pipes  # noqa
    print('pipes imported')
except ImportError:
    failed_modules.append('pipes')

try:
    import PixMapWrapper  # noqa
    print('PixMapWrapper imported')
except ImportError:
    failed_modules.append('PixMapWrapper')

try:
    import pkgutil  # noqa
    print('pkgutil imported')
except ImportError:
    failed_modules.append('pkgutil')

try:
    import platform  # noqa
    print('platform imported')
except ImportError:
    failed_modules.append('platform')

try:
    import plistlib  # noqa
    print('plistlib imported')
except ImportError:
    failed_modules.append('plistlib')

try:
    import popen2  # noqa
    print('popen2 imported')
except ImportError:
    failed_modules.append('popen2')

try:
    import poplib  # noqa
    print('poplib imported')
except ImportError:
    failed_modules.append('poplib')

try:
    import posix  # noqa
    print('posix imported')
except ImportError:
    failed_modules.append('posix')

try:
    import posixfile  # noqa
    print('posixfile imported')
except ImportError:
    failed_modules.append('posixfile')

try:
    import pprint  # noqa
    print('pprint imported')
except ImportError:
    failed_modules.append('pprint')

try:
    import profile  # noqa
    print('profile imported')
except ImportError:
    failed_modules.append('profile')

try:
    import pstats  # noqa
    print('pstats imported')
except ImportError:
    failed_modules.append('pstats')

try:
    import pty  # noqa
    print('pty imported')
except ImportError:
    failed_modules.append('pty')

try:
    import pwd  # noqa
    print('pwd imported')
except ImportError:
    failed_modules.append('pwd')

try:
    import py_compile  # noqa
    print('py_compile imported')
except ImportError:
    failed_modules.append('py_compile')

try:
    import pyclbr  # noqa
    print('pyclbr imported')
except ImportError:
    failed_modules.append('pyclbr')

try:
    import pydoc  # noqa
    print('pydoc imported')
except ImportError:
    failed_modules.append('pydoc')

try:
    import Queue  # noqa
    print('Queue imported')
except ImportError:
    failed_modules.append('Queue')

try:
    import quopri  # noqa
    print('quopri imported')
except ImportError:
    failed_modules.append('quopri')

try:
    import random  # noqa
    print('random imported')
except ImportError:
    failed_modules.append('random')

try:
    import re  # noqa
    print('re imported')
except ImportError:
    failed_modules.append('re')

try:
    import readline  # noqa
    print('readline imported')
except ImportError:
    failed_modules.append('readline')

try:
    import resource  # noqa
    print('resource imported')
except ImportError:
    failed_modules.append('resource')

try:
    import rexec  # noqa
    print('rexec imported')
except ImportError:
    failed_modules.append('rexec')

try:
    import rfc822  # noqa
    print('rfc822 imported')
except ImportError:
    failed_modules.append('rfc822')

try:
    import rlcompleter  # noqa
    print('rlcompleter imported')
except ImportError:
    failed_modules.append('rlcompleter')

try:
    import robotparser  # noqa
    print('robotparser imported')
except ImportError:
    failed_modules.append('robotparser')

try:
    import runpy  # noqa
    print('runpy imported')
except ImportError:
    failed_modules.append('runpy')

try:
    import sched  # noqa
    print('sched imported')
except ImportError:
    failed_modules.append('sched')

try:
    import ScrolledText  # noqa
    print('ScrolledText imported')
except ImportError:
    failed_modules.append('ScrolledText')

try:
    import select  # noqa
    print('select imported')
except ImportError:
    failed_modules.append('select')

try:
    import sets  # noqa
    print('sets imported')
except ImportError:
    failed_modules.append('sets')

try:
    import sgmllib  # noqa
    print('sgmllib imported')
except ImportError:
    failed_modules.append('sgmllib')

try:
    import sha  # noqa
    print('sha imported')
except ImportError:
    failed_modules.append('sha')

try:
    import shelve  # noqa
    print('shelve imported')
except ImportError:
    failed_modules.append('shelve')

try:
    import shlex  # noqa
    print('shlex imported')
except ImportError:
    failed_modules.append('shlex')

try:
    import shutil  # noqa
    print('shutil imported')
except ImportError:
    failed_modules.append('shutil')

try:
    import signal  # noqa
    print('signal imported')
except ImportError:
    failed_modules.append('signal')

try:
    import SimpleHTTPServer  # noqa
    print('SimpleHTTPServer imported')
except ImportError:
    failed_modules.append('SimpleHTTPServer')

try:
    import SimpleXMLRPCServer  # noqa
    print('SimpleXMLRPCServer imported')
except ImportError:
    failed_modules.append('SimpleXMLRPCServer')

try:
    import site  # noqa
    print('site imported')
except ImportError:
    failed_modules.append('site')

try:
    import smtpd  # noqa
    print('smtpd imported')
except ImportError:
    failed_modules.append('smtpd')

try:
    import smtplib  # noqa
    print('smtplib imported')
except ImportError:
    failed_modules.append('smtplib')

try:
    import sndhdr  # noqa
    print('sndhdr imported')
except ImportError:
    failed_modules.append('sndhdr')

try:
    import socket  # noqa
    print('socket imported')
except ImportError:
    failed_modules.append('socket')

try:
    import SocketServer  # noqa
    print('SocketServer imported')
except ImportError:
    failed_modules.append('SocketServer')

try:
    import spwd  # noqa
    print('spwd imported')
except ImportError:
    failed_modules.append('spwd')

try:
    import sqlite3  # noqa
    print('sqlite3 imported')
except ImportError:
    failed_modules.append('sqlite3')

try:
    import ssl  # noqa
    print('ssl imported')
except ImportError:
    failed_modules.append('ssl')

try:
    import stat  # noqa
    print('stat imported')
except ImportError:
    failed_modules.append('stat')

try:
    import statvfs  # noqa
    print('statvfs imported')
except ImportError:
    failed_modules.append('statvfs')

try:
    import string  # noqa
    print('string imported')
except ImportError:
    failed_modules.append('string')

try:
    import StringIO  # noqa
    print('StringIO imported')
except ImportError:
    failed_modules.append('StringIO')

try:
    import stringprep  # noqa
    print('stringprep imported')
except ImportError:
    failed_modules.append('stringprep')

try:
    import struct  # noqa
    print('struct imported')
except ImportError:
    failed_modules.append('struct')

try:
    import subprocess  # noqa
    print('subprocess imported')
except ImportError:
    failed_modules.append('subprocess')

try:
    import sunau  # noqa
    print('sunau imported')
except ImportError:
    failed_modules.append('sunau')

try:
    import sunaudiodev  # noqa
    print('sunaudiodev imported')
except ImportError:
    failed_modules.append('sunaudiodev')

try:
    import SUNAUDIODEV  # noqa
    print('SUNAUDIODEV imported')
except ImportError:
    failed_modules.append('SUNAUDIODEV')

try:
    import symbol  # noqa
    print('symbol imported')
except ImportError:
    failed_modules.append('symbol')

try:
    import symtable  # noqa
    print('symtable imported')
except ImportError:
    failed_modules.append('symtable')

try:
    import sys  # noqa
    print('sys imported')
except ImportError:
    failed_modules.append('sys')

try:
    import sysconfig  # noqa
    print('sysconfig imported')
except ImportError:
    failed_modules.append('sysconfig')

try:
    import syslog  # noqa
    print('syslog imported')
except ImportError:
    failed_modules.append('syslog')

try:
    import tabnanny  # noqa
    print('tabnanny imported')
except ImportError:
    failed_modules.append('tabnanny')

try:
    import tarfile  # noqa
    print('tarfile imported')
except ImportError:
    failed_modules.append('tarfile')

try:
    import telnetlib  # noqa
    print('telnetlib imported')
except ImportError:
    failed_modules.append('telnetlib')

try:
    import tempfile  # noqa
    print('tempfile imported')
except ImportError:
    failed_modules.append('tempfile')

try:
    import termios  # noqa
    print('termios imported')
except ImportError:
    failed_modules.append('termios')

try:
    import test  # noqa
    print('test imported')
except ImportError:
    failed_modules.append('test')

try:
    import textwrap  # noqa
    print('textwrap imported')
except ImportError:
    failed_modules.append('textwrap')

try:
    import thread  # noqa
    print('thread imported')
except ImportError:
    failed_modules.append('thread')

try:
    import threading  # noqa
    print('threading imported')
except ImportError:
    failed_modules.append('threading')

try:
    import time  # noqa
    print('time imported')
except ImportError:
    failed_modules.append('time')

try:
    import timeit  # noqa
    print('timeit imported')
except ImportError:
    failed_modules.append('timeit')

try:
    import Tix  # noqa
    print('Tix imported')
except ImportError:
    failed_modules.append('Tix')

try:
    import Tkinter  # noqa
    print('Tkinter imported')
except ImportError:
    failed_modules.append('Tkinter')

try:
    import token  # noqa
    print('token imported')
except ImportError:
    failed_modules.append('token')

try:
    import tokenize  # noqa
    print('tokenize imported')
except ImportError:
    failed_modules.append('tokenize')

try:
    import trace  # noqa
    print('trace imported')
except ImportError:
    failed_modules.append('trace')

try:
    import traceback  # noqa
    print('traceback imported')
except ImportError:
    failed_modules.append('traceback')

try:
    import ttk  # noqa
    print('ttk imported')
except ImportError:
    failed_modules.append('ttk')

try:
    import tty  # noqa
    print('tty imported')
except ImportError:
    failed_modules.append('tty')

try:
    import turtle  # noqa
    print('turtle imported')
except ImportError:
    failed_modules.append('turtle')

try:
    import types  # noqa
    print('types imported')
except ImportError:
    failed_modules.append('types')

try:
    import unicodedata  # noqa
    print('unicodedata imported')
except ImportError:
    failed_modules.append('unicodedata')

try:
    import unittest  # noqa
    print('unittest imported')
except ImportError:
    failed_modules.append('unittest')

try:
    import urllib  # noqa
    print('urllib imported')
except ImportError:
    failed_modules.append('urllib')

try:
    import urllib2  # noqa
    print('urllib2 imported')
except ImportError:
    failed_modules.append('urllib2')

try:
    import urlparse  # noqa
    print('urlparse imported')
except ImportError:
    failed_modules.append('urlparse')

try:
    import user  # noqa
    print('user imported')
except ImportError:
    failed_modules.append('user')

try:
    import UserDict  # noqa
    print('UserDict imported')
except ImportError:
    failed_modules.append('UserDict')

try:
    import UserList  # noqa
    print('UserList imported')
except ImportError:
    failed_modules.append('UserList')

try:
    import UserString  # noqa
    print('UserString imported')
except ImportError:
    failed_modules.append('UserString')

try:
    import uu  # noqa
    print('uu imported')
except ImportError:
    failed_modules.append('uu')

try:
    import uuid  # noqa
    print('uuid imported')
except ImportError:
    failed_modules.append('uuid')

try:
    import videoreader  # noqa
    print('videoreader imported')
except ImportError:
    failed_modules.append('videoreader')

try:
    import W  # noqa
    print('W imported')
except ImportError:
    failed_modules.append('W')

try:
    import warnings  # noqa
    print('warnings imported')
except ImportError:
    failed_modules.append('warnings')

try:
    import wave  # noqa
    print('wave imported')
except ImportError:
    failed_modules.append('wave')

try:
    import weakref  # noqa
    print('weakref imported')
except ImportError:
    failed_modules.append('weakref')

try:
    import webbrowser  # noqa
    print('webbrowser imported')
except ImportError:
    failed_modules.append('webbrowser')

try:
    import whichdb  # noqa
    print('whichdb imported')
except ImportError:
    failed_modules.append('whichdb')

try:
    import winsound  # noqa
    print('winsound imported')
except ImportError:
    failed_modules.append('winsound')

try:
    import wsgiref  # noqa
    print('wsgiref imported')
except ImportError:
    failed_modules.append('wsgiref')

try:
    import xdrlib  # noqa
    print('xdrlib imported')
except ImportError:
    failed_modules.append('xdrlib')

try:
    import xml  # noqa
    print('xml imported')
except ImportError:
    failed_modules.append('xml')

try:
    import xmlrpclib  # noqa
    print('xmlrpclib imported')
except ImportError:
    failed_modules.append('xmlrpclib')

try:
    import zipfile  # noqa
    print('zipfile imported')
except ImportError:
    failed_modules.append('zipfile')

try:
    import zipimport  # noqa
    print('zipimport imported')
except ImportError:
    failed_modules.append('zipimport')

try:
    import zlib  # noqa
    print('zlib imported')
except ImportError:
    failed_modules.append('zlib')

print('Failed to load these modules:')
print('\n'.join(failed_modules))
