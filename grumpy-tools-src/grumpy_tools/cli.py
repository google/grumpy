# -*- coding: utf-8 -*-

"""Console script for grumpy_tools."""
import sys
from pkg_resources import resource_filename, Requirement, DistributionNotFound

import click

from . import grumpc, grumprun, pydeps


@click.group('grumpy')
def main(args=None):
    """Console script for grumpy_tools."""
    return 0


@main.command('transpile')
@click.argument('script')
@click.option('-m', '-modname', '--modname', default='__main__', help='Python module name')
def transpile(script=None, modname=None):
    """
    Translates the python SCRIPT file to Go, then prints to stdout
    """
    result = grumpc.main(script=script, modname=modname)
    sys.exit(result)


@main.command('run')
@click.option('-m', '-modname', '--modname', help='Run the named module')
def run(modname=None):
    try:
        runtime_gopath = resource_filename(
            Requirement.parse('grumpy-runtime'),
            'grumpy_runtime/data/gopath',
        )
    except DistributionNotFound:
        runtime_gopath = None

    if not runtime_gopath:
        raise click.ClickException("Could not found the Grumpy Runtime 'data/gopath' resource.\n"
                                   "Is 'grumpy-runtime' package installed?")
    result = grumprun.main(modname=modname)
    sys.exit(result)


@main.command('depends')
@click.argument('script')
@click.option('-m', '-modname', '--modname', default='__main__', help='Python module name')
def depends(script=None, modname=None):
    """
    Translates the python SCRIPT file to Go, then prints to stdout
    """
    result = pydeps.main(script=script, modname=modname)
    sys.exit(result)


if __name__ == "__main__":
    import sys
    sys.exit(main())
