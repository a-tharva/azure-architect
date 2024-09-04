from random import randint

from flask import Flask, render_template
app = Flask(__name__, template_folder='templates')

@app.route('/app/')
def index():
    return render_template('index.html')

@app.route('/app/test')
def test():
    return render_template('test.html')

@app.route('/app/info')
def info():
    return render_template('info.html')

@app.route('/app/random')
def random():
    return f'{randint(0,10)}'


if __name__ == '__main__':
    app.run(host="0.0.0.0", port=8000)