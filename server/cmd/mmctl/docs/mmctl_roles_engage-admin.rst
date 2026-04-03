.. _mmctl_roles_engage-admin:

mmctl roles engage-admin
------------------------

Set a user as engage admin

Synopsis
~~~~~~~~


Assign the system_engage_admin role to one or more users. The role is created if it does not yet exist.

::

  mmctl roles engage-admin [users] [flags]

Examples
~~~~~~~~

::

    # Assign engage admin role to a single user
    $ mmctl roles engage-admin john_doe

    # Or assign to multiple users at the same time
    $ mmctl roles engage-admin john_doe jane_doe

Options
~~~~~~~

::

  -h, --help   help for engage-admin

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

* `mmctl roles <mmctl_roles.rst>`_ 	 - Manage user roles

