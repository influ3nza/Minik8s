import requests


def run(username, image, x, y, r, g, b, mark, mysqlIp, rate, threshold1, threshold2, kernel_x, kernel_y, status):

    img_slice = image.split(",")[1]
    img = img_slice.encode("utf-8")
    url = "https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=flmOcOJEQ8mJQGAIydRZWtfo&client_secret=kCyUV5vm0oA8z0DjzwHjofOBPguvAxk0"

    payload = ""
    headers = {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
    }

    response = requests.request("POST", url, headers=headers, data=payload)
    data = response.json()
    access_token = data.get('access_token')
    print(access_token)

    request_url = "https://aip.baidubce.com/rest/2.0/solution/v1/img_censor/v2/user_defined"

    params = {"image": img}
    request_url = request_url + "?access_token=" + access_token
    headers = {'content-type': 'application/x-www-form-urlencoded'}
    response = requests.post(request_url, data=params, headers=headers)
    if response:
        rsp = response.json()
        print(rsp)
        conclusion = rsp.get("conclusion")
        if conclusion == '合规':
            return {
                "username": username,
                "image": image,
                "x": x,
                "y": y,
                "r": r,
                "g": g,
                "b": b,
                "mark": mark,
                "mysqlIp": mysqlIp,
                "rate": rate,
                "threshold1": threshold1,
                "threshold2": threshold2,
                "kernel_x": kernel_x,
                "kernel_y": kernel_y,
                "status": "verify"
            }
        else:
            return {
                "image": "",
                "mark": mark,
                "x": x,
                "y": y,
                "r": r,
                "g": g,
                "b": b,
                "mysqlIp": mysqlIp,
                "rate": rate,
                "threshold1": threshold1,
                "threshold2": threshold2,
                "kernel_x": kernel_x,
                "kernel_y": kernel_y,
                "status": "error in verify"
            }
    else:
        return {
            "image": "",
            "mark": mark,
            "x": x,
            "y": y,
            "r": r,
            "g": g,
            "b": b,
            "mysqlIp": mysqlIp,
            "rate": rate,
            "threshold1": threshold1,
            "threshold2": threshold2,
            "kernel_x": kernel_x,
            "kernel_y": kernel_y,
            "status": "error in verify"
        }
