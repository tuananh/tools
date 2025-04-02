import os
import json
from googleapiclient.errors import HttpError

from helpers import client, fetch_email_or_draft


def main():
    email_id = os.getenv('EMAIL_ID')
    if email_id is None and email_subject is None:
        raise ValueError("Either email_id or email_subject must be set")

    service = client('gmail', 'v1')
    try:
        msg = fetch_email_or_draft(service, email_id)
        if 'payload' not in msg:
            print(json.dumps({'attachments': []}))
            return

        attachments = []
        if 'parts' in msg['payload']:
            for part in msg['payload']['parts']:
                if part.get('filename') and part.get('body', {}).get('attachmentId'):
                    attachments.append({
                        'id': part['body']['attachmentId'],
                        'filename': part['filename']
                    })

        print(json.dumps({'attachments': attachments}))

    except HttpError as error:
        print(json.dumps({'error': str(error)}))


if __name__ == "__main__":
    main()
