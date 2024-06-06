import os

def run(username, status, error):
    log_dir = "log"
    user_log_file = f'{log_dir}/{username}.txt'

    if not os.path.exists(log_dir):
        os.makedirs(log_dir)

    if not os.path.isfile(user_log_file):
        with open(user_log_file, 'w') as file:
            pass

    with open(user_log_file, 'a') as file:
        file.write(f'[Serverless/test] LOG {username}, Status: {status}, Error: {error}\n')


if __name__ == "__main__":
    params = {
        "username": "testUser",
        "status": "testSuccess",
        "error": "testError"
    }

    run(**params)
