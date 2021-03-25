Project description
----------------

Users of an online **document library** are presented with an input form, where they can submit *documents*
(e.g., books, poems, recipes) along with *metadata* (e.g., author, mime type, ISBN).
For the sake of simplicity, they can view *all* stored documents on a single page.

The task was to develop the online **document library** consisting of multiple services.
The application has the following architecture represented below.

Aspects that are used within that project
-----------------------------------------

* How to use Git
* What is Docker, how does it work
* Dockerfiles
* Docker-Compose
* Setup networks in Docker
* Mount volumes in Dockerfile
* Programming in Go
* Fundamental knowledge of Hbase, ZooKeeper, and REST


Links
-----

* `Docker Docs <https://docs.docker.com/>`_
* `Docker Compose getting started <https://docs.docker.com/compose/gettingstarted/>`_
* `Docker Compose file reference <https://docs.docker.com/compose/compose-file/>`_
* `Apache HBase Reference Guide <http://hbase.apache.org/book.html>`_
* `ZooKeeper Documentation <http://zookeeper.apache.org/doc/trunk/>`_
* `Go Documentation <https://golang.org/doc/>`_
* `Pro Git <https://git-scm.com/book/en/v2>`_

Components
----------

In the following, the text provides an overview of the different components.

Nginx
~~~~~

Nginx is a web server that delivers static content in our architecture.
Static content comprises the landing page (index.html), JavaScript, css and font files located in ``nginx/www``.

1. Edit and complete the ``nginx/Dockerfile``

   a) Nginx
   #) Run nginx on container startup

#. Central docker-compose file

   a) Build the image using the Dockerfile for nginx
   #) Assign nginx to the ``se_backend`` network
   #) Mount the host directory ``nginx/www`` to ``/var/www/nginx`` in the container

#. Verify your setup (it should display the landing page)

HBase
~~~~~

We use HBase, the open source implementation of Bigtable, as database.
``hbase/hbase_init.txt`` creates the ``se2`` namespace and a ``library`` table with two column families: ``document`` and ``metadata``.

1. Build the docker image for the Dockerfile located in ``hbase/``
#. Edit the docker-compose file
   
   * Add hbase to the ``se_backend`` network

#. Start the hbase container to test HBase:

   * The Container exposes different ports for different APIs.
   * We recommend to use the JSON REST API, but choose whatever API suits you best.
   * `HBase REST documentation <http://hbase.apache.org/book.html#_rest>`_
   * The client port for REST is 8080
   * Use Curl to explore the API
      * ``curl -vi -X PUT -H "Content-Type: application/json" -d '<json row description>' "localhost:8080/se2:library/fakerow"``
      * yes, it's really *fakerow*
   
ZooKeeper
~~~~~~~~~

Deviating from the architecture image, you don't need to create an extra ZooKeeper container.
**The HBase image above already contains a ZooKeeper instance.**

1. Add an alias to the hbase section in the docker-compose file such that other containers can connect to it by referring to the name ``zookeeper``


Grproxy
~~~~~~~

This is the first service/server.
There is an implemented reverse proxy that forwards every request to nginx, except those with a "library" prefix in the path (e.g., ``http://host/library``).
It is need also to discover running gserve instances with the help of teh ZooKeeper service and forward ``library`` requests in circular order among those instances (Round Robin).

1. The reverse proxy implemented in *grproxy/src/grproxy/grproxy.go*
#. Dockerfile ``grproxy/Dockerfile`` is created
#. In the docker-compose file:

   a) Built the grproxy container image
   #) Added grproxy to both networks: ``se_frontend`` and ``se_backend``


Gserve
~~~~~~

Gserve is the second service and it serves two purposes.
Firstly, it receives ``POST`` requests from the client (via grproxy) and adds or alters rows in HBase.
And secondly, it replies to ``GET`` requests with an HTML page displaying the contents of the whole document library.
It only receives requests from grproxy after it subscribed to ZooKeeper, and automatically unsubscribes from ZooKeeper if it shuts down or crashes.

1. Gserve returns all versions of HBase cells (see output sample above)
#. The returned HTML page contains the string *"proudly served by gserve1"* (or gserve2, ...) without HTML tags in between
#. In the docker-compose file

   a) Built the gserve container
   #) Started two instances *gserve1* and *gserve2*
   #) Added both instances to the ``se_backend`` network
   #) Both instances start after hbase and grproxy
   #) Provided the names of the instances (gserve1, gserve2) via environment variables