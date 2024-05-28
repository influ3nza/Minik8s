import base64
from PIL import Image, ImageDraw, ImageFont
from io import BytesIO


def run(username, image, x, y, r, g, b, mark, mysqlIp, status):
    color = (r, g, b)
    print(type(color))
    img_slice = image.split(",")[1]
    img_b64decoded = base64.b64decode(img_slice)
    image_ = Image.open(BytesIO(img_b64decoded))
    image_ = image_.convert("RGBA")
    draw = ImageDraw.Draw(image_)
    try:
        font = ImageFont.truetype("arial.ttf", 40)  # 使用系统字体
    except IOError:
        font = ImageFont.load_default()

    draw.text((x, y), mark, color, font=font)

    buffered = BytesIO()

    image_.save(buffered, format="PNG")
    image_.save('watermarked_image.png')

    im_base64 = base64.b64encode(buffered.getvalue()).decode('utf-8')

    value = f"data:image/png;base64,{im_base64}"

    image_data = base64.b64decode(im_base64)
    with open("water.png", "wb") as f:
        f.write(image_data)

    with open("waterb64", "wb") as file:
        file.write(value.encode('utf-8'))


    return {
        "username": username,
        "image": value,
        "mysqlIp": mysqlIp,
        "status": "marked"
    }

def main():
    file_path = "./edgeb64"
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
        "mark": "I'am ML ing",
        "mysqlIp": "10.2.1.25",
        "status": "start"
    }

    run(**params)

if __name__ == '__main__':
    main()
