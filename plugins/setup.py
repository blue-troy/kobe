# -*- coding: UTF-8 -*-
import os

import setuptools

setuptools.setup(
    name='ansible-kobe-plugin',
    version='0.0.6',
    keywords='kobe',
    description='A plugin for ansible to connect kobe',
    packages=setuptools.find_packages(exclude=['plugins/*']),
    license='MIT',
    data_files=[
        ('/var/kobe/lib/ansible/plugins/callback', ['plugins/callback/result.py']),
    ],
    include_package_data=True,
    install_requires=["grpcio", "grpcio-tools", "ansible"]
)
