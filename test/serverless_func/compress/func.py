import base64
import numpy as np
import cv2


def run(username, image, x, y, r, g, b, mark, mysqlIp, rate, threshold1, threshold2, kernel_x, kernel_y, status):
    img_slice = image.split(",")[1]
    img_b64decoded = base64.b64decode(img_slice)
    nparr = np.frombuffer(img_b64decoded, np.uint8)
    img_cv = cv2.imdecode(nparr, cv2.IMREAD_COLOR)

    print(img_cv.shape)
    width = int(img_cv.shape[1] * rate)
    height = int(img_cv.shape[0] * rate)
    dim = (width, height)

    resized_img = cv2.resize(img_cv, dim, interpolation=cv2.INTER_AREA)
    _, im_arr = cv2.imencode('.png', resized_img)  # im_arr: 图像转换为numpy数组
    im_bytes = im_arr.tobytes()

    im_base64 = base64.b64encode(im_bytes).decode('utf-8')
    value = f"data:image/png;base64,{im_base64}"

    image_data = base64.b64decode(im_base64)

    with open("../compressed.png", "wb") as f:
        f.write(image_data)

    with open('../compressedb64', "wb") as file:
        file.write(value.encode('utf-8'))
    # print(im_base64)

    return {
        "username": username,
        "image": value,
        "x": x,
        "y": y,
        "r": r,
        "g": g,
        "b": b,
        "mark": mark,
        "mysqlIp": mysqlIp,
        "threshold1": threshold1,
        "threshold2": threshold2,
        "kernel_x": kernel_x,
        "kernel_y": kernel_y,
        "status": "compressed"
    }


def main():
    file_path = "../testimg"
    with open(file_path, "rb") as file:
        file_content = file.read()

    val = file_content.decode('utf-8')
    params = {
        "username": "testuser",
        "image": val,
        "x": 20,
        "y": 20,
        "r": 255,
        "g": 255,
        "b": 255,
        "rate": 0.9,
        "mark": "water_mark",
        "mysqlIp": "10.2.1.25",
        "threshold1": 50,
        "threshold2": 150,
        "kernel_x": 3,
        "kernel_y": 3,
        "status": "start"
    }

    run(**params)


if __name__ == '__main__':
    main()
