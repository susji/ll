# ll

`ll` is a minimalistic URL shortener server. You probably want to run
it behind some TLS-terminating reverse proxy.

## usage

    $ ll -h

## deploying

	$ cd ll
	$ make
	$ python3 -m venv ~/venv-pyinfra
	$ ~/venv-pyinfra/bin/pip install pyinfra
	$ source ~/venv-pyinfra/bin/activate
	$ cd deploy
	$ cp misc/ll.conf.example ll.conf
	$ [edit ll.conf your liking]
	$ pyinfra --data LL_DOMAIN=a.example.com a.example.com deploy.py
