from random import randint
from flask import Flask, render_template, jsonify
import grpc
import prime_pb2
import prime_pb2_grpc

app = Flask(__name__, template_folder='templates')

@app.route('/api/')
def index():
    return render_template('index.html')

@app.route('/api/test')
def test():
    return render_template('test.html')

@app.route('/api/info')
def info():
    return render_template('info.html')

@app.route('/api/random')
def random():
    return f'{randint(0,10)}'

@app.route('/api/prime/<path:nums>')
def prime(nums):
    try:
        ns = [int(x) for x in nums.split(',')]
    except ValueError:
        return "Invalid input: please provide comma-separated integers (e.g., 10,12,14)", 400

    # Use Kubernetes service name for gRPC
    channel = grpc.insecure_channel('go-prime-service.default.svc.cluster.local:50051')
    stub = prime_pb2_grpc.PrimeServiceStub(channel)

    request = prime_pb2.PrimeRequest(ns=ns)
    try:
        response = stub.GetPrimes(request)
        return jsonify({"primes": list(response.primes)})
    except grpc.RpcError as e:
        return f"gRPC error: {str(e)}", 500

if __name__ == '__main__':
    app.run(host="0.0.0.0", port=8000) 