import base64
import numpy as np
import cv2

def run(username, image, x, y, r, g, b, mark, mysqlIp, threshold1, threshold2, kernel_x, kernel_y, status):
    base64_str = image.split(",")[1]
    img_data = base64.b64decode(base64_str)
    nparr = np.fromstring(img_data, np.uint8)
    img_cv = cv2.imdecode(nparr, cv2.IMREAD_COLOR)

    gray = cv2.cvtColor(img_cv, cv2.COLOR_BGR2GRAY)
    
    blurred_img = cv2.GaussianBlur(gray, (kernel_x, kernel_y), 0)

    edges = cv2.Canny(blurred_img, threshold1=threshold1, threshold2=threshold2)

    _, im_arr = cv2.imencode('.png', edges)
    im_bytes = im_arr.tobytes()
    im_base64 = base64.b64encode(im_bytes).decode('utf-8')
    value = f"data:image/png;base64,{im_base64}"

    image_data = base64.b64decode(im_base64)

    with open("edged.png", "wb") as f:
        f.write(image_data)

    with open("edgeb64", "wb") as file:
        file.write(value.encode('utf-8'))

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
        "status": "edgeFinished"
    }

def main():
    file_path = "./compressedb64"
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