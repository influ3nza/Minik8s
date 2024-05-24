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
        response = Response(json.dumps(result), headers=headers, status=200)
        return response
    except TypeError as e:
        err = {
            "err": e
        }
        response = Response(json.dumps(err), headers=headers, status=200)
        return response
    except Exception as e:
        response = Response(json.dumps(err), headers=headers, status=500)
        return response


if __name__ == '__main__':
    app.run(host="0.0.0.0", port=8080, debug=True)
