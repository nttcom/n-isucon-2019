from datetime import datetime as dt
import subprocess
import secrets


def get_salt():
    return secrets.token_hex(64)


def get_passwordhash(salt, password):
    string = password + salt
    for _ in range(1000):
        cmd = 'echo -n ' + string + ' | openssl sha256'
        string = (subprocess.check_output(cmd, shell=True)
                            .decode('utf-8')
                            .partition(' ')[2]
                            .rstrip())
    return string


def get_today():
    now = dt.today()
    result = now.strftime('%Y-%m-%d %H:%M:%S')
    return result
