PIOT Golang
===========

.. image:: https://dev.azure.com/michalnezerka/PIOT/_apis/build/status/mnezerka.go-piot?branchName=master

Golang package provides services and utilities for PIOT infrastructure tools.


Development environment - minimal
---------------------------------

1. Run mongodb docker container::

     docker-compose up -d mongodb

2. Run script ``scripts/env.sh`` to get IP address of mongo container
   and set env variable for piot server

3. Run tests (not in parallel since shared mongodb is used)::

     # all tests
     go test

     # tests for selected test case (matched against regexp)
     go test --run ShortNotation
