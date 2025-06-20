import firebase_admin
from firebase_admin import credentials
from firebase_admin import firestore

cred = credentials.Certificate('serviceAccount.json')

class FireStore:
    def __init__(self):
        app = firebase_admin.initialize_app(cred)
        self.db = firestore.client()

    def read(self, collection_name):
        docs = self.db.collection(collection_name).stream()
        data = []
        for doc in docs:
            data.append(doc.to_dict())
        return data