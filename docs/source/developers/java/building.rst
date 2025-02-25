.. Licensed to the Apache Software Foundation (ASF) under one
.. or more contributor license agreements.  See the NOTICE file
.. distributed with this work for additional information
.. regarding copyright ownership.  The ASF licenses this file
.. to you under the Apache License, Version 2.0 (the
.. "License"); you may not use this file except in compliance
.. with the License.  You may obtain a copy of the License at

..   http://www.apache.org/licenses/LICENSE-2.0

.. Unless required by applicable law or agreed to in writing,
.. software distributed under the License is distributed on an
.. "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
.. KIND, either express or implied.  See the License for the
.. specific language governing permissions and limitations
.. under the License.

.. highlight:: console

.. _building-arrow-java:

===================
Building Arrow Java
===================

.. contents::

System Setup
============

Arrow Java uses the `Maven <https://maven.apache.org/>`_ build system.

Building requires:

* JDK 8, 9, 10, 11, 17, or 18, but only JDK 8, 11 and 17 are tested in CI.
* Maven 3+

Building
========

All the instructions below assume that you have cloned the Arrow git
repository:

.. code-block::

    $ git clone https://github.com/apache/arrow.git
    $ cd arrow
    $ git submodule update --init --recursive

Basic Installation
------------------

To build the default modules, go to the project root and execute:

.. code-block::

    $ cd arrow/java
    $ export JAVA_HOME=<absolute path to your java home>
    $ java --version
    $ mvn clean install

Building JNI Libraries on Linux
-------------------------------

First, we need to build the `C++ shared libraries`_ that the JNI bindings will use.
We can build these manually or we can use `Archery`_ to build them using a Docker container
(This will require installing Docker, Docker Compose, and Archery).

.. code-block::

    $ cd arrow
    $ archery docker run java-jni-manylinux-2014
    $ ls -latr java-dist/
    |__ libarrow_cdata_jni.so
    |__ libarrow_dataset_jni.so
    |__ libarrow_orc_jni.so
    |__ libgandiva_jni.so

Building JNI Libraries on MacOS
-------------------------------
Note: If you are building on Apple Silicon, be sure to use a JDK version that was compiled for that architecture. See, for example, the `Azul JDK <https://www.azul.com/downloads/?os=macos&architecture=arm-64-bit&package=jdk>`_.

To build only the C Data Interface library:

.. code-block::

    $ cd arrow
    $ brew bundle --file=cpp/Brewfile
    Homebrew Bundle complete! 25 Brewfile dependencies now installed.
    $ export JAVA_HOME=<absolute path to your java home>
    $ mkdir -p java-dist java-native-c
    $ cd java-native-c
    $ cmake \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_INSTALL_PREFIX=../java-dist/lib \
        ../java
    $ cmake --build . --target install
    $ ls -latr ../java-dist/lib
    |__ libarrow_cdata_jni.dylib

To build other JNI libraries:

.. code-block::

    $ cd arrow
    $ brew bundle --file=cpp/Brewfile
    Homebrew Bundle complete! 25 Brewfile dependencies now installed.
    $ export JAVA_HOME=<absolute path to your java home>
    $ mkdir -p java-dist java-native-cpp
    $ cd java-native-cpp
    $ cmake \
        -DARROW_BOOST_USE_SHARED=OFF \
        -DARROW_BROTLI_USE_SHARED=OFF \
        -DARROW_BZ2_USE_SHARED=OFF \
        -DARROW_GFLAGS_USE_SHARED=OFF \
        -DARROW_GRPC_USE_SHARED=OFF \
        -DARROW_LZ4_USE_SHARED=OFF \
        -DARROW_OPENSSL_USE_SHARED=OFF \
        -DARROW_PROTOBUF_USE_SHARED=OFF \
        -DARROW_SNAPPY_USE_SHARED=OFF \
        -DARROW_THRIFT_USE_SHARED=OFF \
        -DARROW_UTF8PROC_USE_SHARED=OFF \
        -DARROW_ZSTD_USE_SHARED=OFF \
        -DARROW_JNI=ON \
        -DARROW_PARQUET=ON \
        -DARROW_FILESYSTEM=ON \
        -DARROW_DATASET=ON \
        -DARROW_GANDIVA_JAVA=ON \
        -DARROW_GANDIVA_STATIC_LIBSTDCPP=ON \
        -DARROW_GANDIVA=ON \
        -DARROW_ORC=ON \
        -DARROW_PLASMA_JAVA_CLIENT=ON \
        -DARROW_PLASMA=ON \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_INSTALL_LIBDIR=lib \
        -DCMAKE_INSTALL_PREFIX=../java-dist \
        -DCMAKE_UNITY_BUILD=ON \
        -Dre2_SOURCE=BUNDLED \
        -DBoost_SOURCE=BUNDLED \
        -Dutf8proc_SOURCE=BUNDLED \
        -DSnappy_SOURCE=BUNDLED \
        -DORC_SOURCE=BUNDLED \
        -DZLIB_SOURCE=BUNDLED \
        ../cpp
    $ cmake --build . --target install
    $ ls -latr  ../java-dist/lib
    |__ libarrow_dataset_jni.dylib
    |__ libarrow_orc_jni.dylib
    |__ libgandiva_jni.dylib

Building Arrow JNI Modules
--------------------------

To compile the JNI bindings, use the ``arrow-c-data`` Maven profile:

.. code-block::

    $ cd arrow/java
    $ mvn -Darrow.c.jni.dist.dir=<absolute path to your arrow folder>/java-dist/lib -Parrow-c-data clean install

To compile the JNI bindings for ORC / Gandiva / Dataset, use the ``arrow-jni`` Maven profile:

.. code-block::

    $ cd arrow/java
    $ mvn -Darrow.cpp.build.dir=<absolute path to your arrow folder>/java-dist/lib -Parrow-jni clean install

IDE Configuration
=================

IntelliJ
--------

To start working on Arrow in IntelliJ: build the project once from the command
line using ``mvn clean install``. Then open the ``java/`` subdirectory of the
Arrow repository, and update the following settings:

* In the Files tool window, find the path ``vector/target/generated-sources``,
  right click the directory, and select Mark Directory as > Generated Sources
  Root. There is no need to mark other generated sources directories, as only
  the ``vector`` module generates sources.
* For JDK 8, disable the ``error-prone`` profile to build the project successfully.
* For JDK 11, due to an `IntelliJ bug
  <https://youtrack.jetbrains.com/issue/IDEA-201168>`__, you must go into
  Settings > Build, Execution, Deployment > Compiler > Java Compiler and disable
  "Use '--release' option for cross-compilation (Java 9 and later)". Otherwise
  you will get an error like "package sun.misc does not exist".
* You may want to disable error-prone entirely if it gives spurious
  warnings (disable both error-prone profiles in the Maven tool window
  and "Reload All Maven Projects").
* If using IntelliJ's Maven integration to build, you may need to change
  ``<fork>`` to ``false`` in the pom.xml files due to an `IntelliJ bug
  <https://youtrack.jetbrains.com/issue/IDEA-278903>`__.

You may not need to update all of these settings if you build/test with the
IntelliJ Maven integration instead of with IntelliJ directly.

Common Errors
=============

* When working with the JNI code: if the C++ build cannot find dependencies, with errors like these:

  .. code-block::

     Could NOT find Boost (missing: Boost_INCLUDE_DIR system filesystem)
     Could NOT find Lz4 (missing: LZ4_LIB)
     Could NOT find zstd (missing: ZSTD_LIB)

  Specify that the dependencies should be downloaded at build time (more details at `Dependency Resolution`_):

  .. code-block::

     -Dre2_SOURCE=BUNDLED \
     -DBoost_SOURCE=BUNDLED \
     -Dutf8proc_SOURCE=BUNDLED \
     -DSnappy_SOURCE=BUNDLED \
     -DORC_SOURCE=BUNDLED \
     -DZLIB_SOURCE=BUNDLED

.. _Archery: https://github.com/apache/arrow/blob/master/dev/archery/README.md
.. _Dependency Resolution: https://arrow.apache.org/docs/developers/cpp/building.html#individual-dependency-resolution
.. _C++ shared libraries: https://arrow.apache.org/docs/cpp/build_system.html
