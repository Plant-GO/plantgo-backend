import os
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '3'

import sys
import tensorflow as tf
import numpy as np
from tensorflow.keras.models import load_model
from io import BytesIO
from PIL import Image
import base64

model_path = "ml/flower3.keras"
class_names = ['Marigold', 'Scarlet Sage']

def load_and_preprocess_image_from_bytes(img_bytes, target_size=(150, 150)):
    img = Image.open(BytesIO(img_bytes)).convert('RGB')
    img = img.resize(target_size)
    img_array = np.array(img) / 255.0
    return np.expand_dims(img_array, axis=0)

def main():
    if len(sys.argv) < 2:
        print("Error|0.0")
        sys.exit(1)
        
    b64_file_path = sys.argv[1]

    with open(b64_file_path, "r") as f:
        b64_string = f.read().strip()

    img_bytes = base64.b64decode(b64_string)

    model = load_model(model_path)

    img_array = load_and_preprocess_image_from_bytes(img_bytes)

    preds = model.predict(img_array, verbose=0)

    predicted_index = np.argmax(preds, axis=1)[0]
    predicted_label = class_names[predicted_index]
    confidence = float(np.max(preds))

    print(f"{predicted_label}|{confidence:.4f}")

if __name__ == "__main__":
    main()
