import mysql.connector


def run(username, image, x, y, r, g, b, mark, mysqlIp, rate, threshold1, threshold2, kernel_x, kernel_y, status):
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


def main():
    data = {
        "username": "testuser",
        "image": "nothing",
        "x": 20,
        "y": 20,
        "r": 255,
        "g": 255,
        "b": 255,
        "mark": "water_mark",
        "mysqlIp": "10.2.1.24",
        "rate": 0.8,
        "threshold1": 50,
        "threshold2": 150,
        "kernel_x": 3,
        "kernel_y": 3,
        "status": "start"
    }
    run(**data)


if __name__ == '__main__':
    main()
