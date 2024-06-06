import mysql.connector


def run(username, image, mysqlIp, status):
    db = mysql.connector.connect(
        host=mysqlIp,  # MySQL服务器地址
        user="root",  # 用户名
        password="123456",  # 密码
        database="saveimg"  # 数据库名称
    )
    cursor = db.cursor()
    sql = "INSERT INTO response(image, username) VALUES (%s, %s);"
    cursor.execute(sql, (image, username))

    db.commit()
    cursor.close()
    db.close()
    return {
        "status": "finished",
        "username": username,
        "error": "null"
    }


def main():

    file_path = "./waterb64"
    with open(file_path, "rb") as file:
        file_content = file.read()

    val = file_content.decode('utf-8')
    
    data = {
        "username": "testuser",
        "image": val,
        "mysqlIp": "10.2.1.24",
        "status": "start"
    }
    run(**data)


if __name__ == '__main__':
    main()
