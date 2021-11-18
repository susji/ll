# deploying

	$ cd sp
	$ make
	$ python3 -m venv ~/venv-pyinfra
	$ ~/venv-pyinfra/bin/pip install pyinfra
	$ source ~/venv-pyinfra/bin/activate
	$ cd deploy
	$ cp misc/ll.conf.example ll.conf
	$ [edit ll.conf your liking]
	$ pyinfra --data LL_DOMAIN=a.example.com a.example.com deploy.py
