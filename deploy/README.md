# deploying

    $ scp * ${server}:
    $ ssh -t ${server} \
		'export YOUR_DOMAIN=yourdomain; sudo --preserve-env=YOUR_DOMAIN /bin/sh $HOME/x.sh'
