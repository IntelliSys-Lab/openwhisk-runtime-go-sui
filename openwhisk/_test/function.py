import datetime, json, os, uuid, sys
import warnings
warnings.filterwarnings("ignore")


def print_time():
    print(datetime.datetime.now(), flush=True)



def main():
    print_time()  # print time at the start
    print(sys.executable)
    from PIL import Image
    import torch
    from torchvision import transforms
    from torchvision.models import resnet18
    import boto3

    print_time()  # print time at the start

    SCRIPT_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__)))
    class_idx = json.load(open(os.path.join(SCRIPT_DIR, "imagenet_class_index.json"), 'r'))
    idx2label = [class_idx[str(k)][1] for k in range(len(class_idx))]




    model_bucket = 'test-data-sui'
    input_bucket = 'test-data-sui'
    key = 'beach.jpg'
    model_key1 = 'resnet18.pth'
    print(SCRIPT_DIR)

    access_key_id = 'AKIAV3GBJPI4VX4UGGK6'
    secret_access_key = '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'

    s3 = boto3.client('s3', aws_access_key_id=access_key_id, aws_secret_access_key=secret_access_key)

    #image_path = '/Users/suiyifan/Downloads/serverless-benchmarks-master-2/analyzer/beach.jpg'
    image_path = 'beach.jpg'

    global model
    model = None

    if not model:

        # First Download of Model
        print("start download")

        #model_path = '/Users/suiyifan/Downloads/serverless-benchmarks-master-2/analyzer/sharedmemory/resnet152.pth'
        model_path = 'resnet18.pth'


        print(model_path)

        model_process_begin = datetime.datetime.now()
        model = resnet18(pretrained=False)
        model.load_state_dict(torch.load(model_path))
        model.eval()
        model_process_end = datetime.datetime.now()

    else:

        print("has been downloaded")

        model_process_begin = datetime.datetime.now()
        model_process_end = datetime.datetime.now()
        model_download_begin = datetime.datetime.now()
        model_download_end = model_download_begin

    print_time()  # print time at the start

    line = sys.stdin.readline()  # block until some input is received
    print_time()


    process_begin = datetime.datetime.now()
    input_image = Image.open(image_path)
    preprocess = transforms.Compose([
        transforms.Resize(256),
        transforms.CenterCrop(224),
        transforms.ToTensor(),
        transforms.Normalize(mean=[0.485, 0.456, 0.406], std=[0.229, 0.224, 0.225]),
    ])


    input_batch = preprocess(input_image).unsqueeze(0)  # create a mini-batch as expected by the model
    output = model(input_batch)


    _, index = torch.max(output, 1)
    # The output has unnormalized scores. To get probabilities, you can run a softmax on it.
    prob = torch.nn.functional.softmax(output[0], dim=0)
    _, indices = torch.sort(output, descending=True)
    ret = idx2label[index]
    process_end = datetime.datetime.now()


    #model_download_time = (model_download_end - model_download_begin) / datetime.timedelta(microseconds=1)
    model_process_time = (model_process_end - model_process_begin) / datetime.timedelta(microseconds=1)
    process_time = (process_end - process_begin) / datetime.timedelta(microseconds=1)

    print_time()

    return {
        'result': {'idx': index.item(), 'class': ret},
        'measurement': {
            'image_download_time': 0,
            'model_download_time':0,
            'model_process_time': model_process_time,
            'compute_time': process_time,
            'path':image_path
        }
    }

if __name__ == "__main__":
    print(main())

