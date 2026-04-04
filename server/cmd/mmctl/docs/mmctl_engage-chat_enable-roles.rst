.. _mmctl_engage-chat_enable-roles:

mmctl engage-chat enable-roles
------------------------------

Enable engage-chat custom roles

Synopsis
~~~~~~~~


Create or restore the engage-chat custom roles (e.g. system_engage_admin).

::

  mmctl engage-chat enable-roles [flags]

Examples
~~~~~~~~

::

    $ mmctl engage-chat enable-roles

Options
~~~~~~~

::

  -h, --help   help for enable-roles

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config string                path to the configuration file (default "$XDG_CONFIG_HOME/mmctl/config")
      --disable-pager                disables paged output
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --json                         the output format will be in json format
      --local                        allows communicating with the server through a unix socket
      --quiet                        prevent mmctl to generate output for the commands
      --strict                       will only run commands if the mmctl version matches the server one
      --suppress-warnings            disables printing warning messages

SEE ALSO
~~~~~~~~

* `mmctl engage-chat <mmctl_engage-chat.rst>`_ 	 - Management of engage-chat features

