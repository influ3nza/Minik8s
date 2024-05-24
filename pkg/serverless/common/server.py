import json

from flask import Flask, request, Response, jsonify
import func

app = Flask(__name__)


@app.route('/', methods=['POST'])
def handle_request():
    # `params` is a dict
    params = request.json
    headers = {'Content-Type': 'application/json'}
    try:
        result = func.run(**params)
        response = jsonify(result)
        response.headers = headers
        response.status = 200
        return response
    except TypeError as e:
        response = jsonify({"error": str(e)})
        response.status_code = 400
        response.headers = headers
        return response
    except Exception as e:
        response = jsonify({"error": str(e)})
        response.status_code = 500
        response.headers = headers
        return response


if __name__ == '__main__':
    app.run(host="0.0.0.0", port=8080, debug=True)
