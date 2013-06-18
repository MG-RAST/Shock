#!/usr/bin/env python

from distutils.core import setup
import shock

setup(name='shock',
      version=shock.__version__,
      author=shock.__author__,
      license=shock.__licence__,
      download_url='https://github.com/MG-RAST/Shock',
      packages=['shock'],
     )