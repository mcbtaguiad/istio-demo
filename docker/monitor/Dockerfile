FROM python:3.14-trixie

LABEL Maintainer="Mark Taguiad <marktaguiad@marktaguiad.dev>"

WORKDIR /app

COPY app.py /app/app.py

RUN pip install --no-cache-dir flask gunicorn

EXPOSE 8000

# CMD ["python", "app.py"]
CMD ["gunicorn", "-w", "4", "-b", "0.0.0.0:8000", "app:app"]
