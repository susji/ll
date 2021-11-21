# ll

`ll` is a minimalistic URL shortener server. Presently it's mainly a
proof-of-concept, so I advise against running it in production. In any
case, you probably want to run it behind some TLS-terminating reverse
proxy, so nginx stuff is included in the example deployment file.

## usage

    $ ll -h

## deploying

	$ cd ll
	$ make
	$ python3 -m venv ~/venv-pyinfra
	$ ~/venv-pyinfra/bin/pip install -r deploy/requirements.txt
	$ source ~/venv-pyinfra/bin/activate
	$ cd deploy
	$ cp ll.conf.example ll.conf       # edit ll.conf to your liking
	$ pyinfra --data LL_DOMAIN=a.example.com a.example.com deploy.py

## todo

- [ ] actual packaging like `.deb` building
