from tqdm import tqdm
from typing import List, Tuple
import multiprocessing
import argparse
import pandas as pd
import datetime
import os
import collections
import contextlib
from posixpath import basename, split
import sys
import wave
from pandas.core import base
import webrtcvad
from csv import reader
import shutil

import subprocess
# ffmpeg = "/Library/ffmpeg"


path_to_input = "./storage"
path_to_tran = "./transition-folder"
path_to_output = "./output-folder"
path_to_result = "./result-folder"


def write_csv(file_name, side, start, stop, token_fileID):
    token = token_fileID.split(".")[0]
    channel = []
    for i in range(len(start)):
        channel.append(side)
    data = {
        'StartTime': start, 'StopTime': stop, 'Channel': channel}

    df = pd.DataFrame.from_dict(data, orient='index')
    df = df.transpose()
    df = pd.DataFrame(data, columns=['StartTime', 'StopTime', 'Channel'])
    # df.to_csv(config.vad_users + "/" + file_name +
    #           '.csv', index=False, header=True)
    df.to_csv(path_to_output+"/"+token + "/" + file_name + "/" + side +
              '.csv', index=False, header=True)
    # os.remove(path_to_tran+"/"+token+"/"+file_name + "/" + side +
    #           ".wav")


def write_result_csv(token_fileID):
    token = token_fileID.split(".")[0]
    fileID = token_fileID.split(".")[1]
    # filenames = list of side [psychotherapist,patient]
    # name_of_file = name of input file
    for path, _, filenames in os.walk(path_to_output + "/" + token):
        if ".DS_Store" in filenames:
            filenames.remove(".DS_Store")
        print(path)

        if (len(filenames) == 2) & (str(path[-9:]) == str(fileID)):
            # name_of_file = path.split("/")[-1]
            name_of_file = os.path.basename(path)
            path_to_file = path_to_output+"/"+token + "/" + name_of_file + "/"
            resultlist = result(
                path_to_file+filenames[0], path_to_file+filenames[1])
            channel_list = []
            start_list = []
            stop_list = []
            for i in range(len(resultlist)):
                channel_list.append(resultlist[i][2])
                start_list.append(resultlist[i][0])
                stop_list.append(resultlist[i][1])
            data = {'Channel': channel_list,
                    'StartTime': start_list, 'StopTime': stop_list}

            df = pd.DataFrame.from_dict(data, orient='index')
            df = df.transpose()
            df = pd.DataFrame(
                data, columns=['Channel', 'StartTime', 'StopTime'])
            if not os.path.exists(path_to_result + "/"+token):
                os.mkdir(path_to_result + "/"+token)
            df.to_csv(path_to_result+"/"+token + "/"+name_of_file +
                      '.csv', index=False, header=True)
            # os.remove(path_to_output+"/"+token+"/" +
            #           name_of_file + "/" + "psychotherapist.csv")
            # os.remove(path_to_output+"/"+token+"/" +
            #           name_of_file + "/" + "patient.csv")

    print("--------- created CSV ---------")


def result(path_channel1, path_channel2):
    list1 = read(path_channel1)
    list2 = read(path_channel2)
    resultlist = []
    for item in list1:
        resultlist.append(item)
    for item in list2:
        resultlist.append(item)
    resultlist.sort()
    return resultlist


def read(filename):
    with open(filename, 'r') as read_obj:
        # pass the file object to reader() to get the reader object
        csv_reader = reader(read_obj)
        # Pass reader object to list() to get a list of lists
        list_of_rows = list(csv_reader)
        if ['StartTime', 'StopTime', 'Channel'] in list_of_rows:
            list_of_rows.remove(['StartTime', 'StopTime', 'Channel'])
        return list_of_rows


def read_wave(path):
    with contextlib.closing(wave.open(path, 'rb')) as wf:
        num_channels = wf.getnchannels()
        assert num_channels == 1
        sample_width = wf.getsampwidth()
        assert sample_width == 2
        sample_rate = wf.getframerate()
        assert sample_rate in (8000, 16000, 32000)
        pcm_data = wf.readframes(wf.getnframes())
        return pcm_data, sample_rate


def write_wave(path, audio, sample_rate):
    with contextlib.closing(wave.open(path, 'wb')) as wf:
        wf.setnchannels(1)
        wf.setsampwidth(2)
        wf.setframerate(sample_rate)
        wf.writeframes(audio)


class Frame(object):
    def __init__(self, bytes, timestamp, duration):
        self.bytes = bytes
        self.timestamp = timestamp
        self.duration = duration


def frame_generator(frame_duration_ms, audio, sample_rate):
    n = int(sample_rate * (frame_duration_ms / 1000.0) * 2)
    offset = 0
    timestamp = 0.0
    duration = (float(n) / sample_rate) / 2.0
    while offset + n < len(audio):
        yield Frame(audio[offset:offset + n], timestamp, duration)
        timestamp += duration
        offset += n


def vad_collector(file_name, side, sample_rate, frame_duration_ms, padding_duration_ms, vad, frames, token_fileID):
    num_padding_frames = int(padding_duration_ms / frame_duration_ms)
    ring_buffer = collections.deque(maxlen=num_padding_frames)
    triggered = False
    voiced_frames = []

    start_time = ''
    stop_time = ''
    vad_start = []
    vad_stop = []

    for frame in frames:
        sys.stdout.write(
            '' if vad.is_speech(frame.bytes, sample_rate) else '')
        if not triggered:
            ring_buffer.append(frame)
            num_voiced = len([f for f in ring_buffer
                              if vad.is_speech(f.bytes, sample_rate)])
            if num_voiced > 0.9 * ring_buffer.maxlen:
                start_time = datetime.timedelta(
                    seconds=(ring_buffer[0].timestamp))
                vad_start.append(str(start_time))

                triggered = True
                voiced_frames.extend(ring_buffer)
                ring_buffer.clear()
        else:
            voiced_frames.append(frame)
            ring_buffer.append(frame)
            num_unvoiced = len([f for f in ring_buffer
                                if not vad.is_speech(f.bytes, sample_rate)])
            if num_unvoiced > 0.9 * ring_buffer.maxlen:
                stop_time = datetime.timedelta(
                    seconds=(frame.timestamp + frame.duration))
                vad_stop.append(str(stop_time))

                triggered = False
                yield b''.join([f.bytes for f in voiced_frames])
                ring_buffer.clear()
                voiced_frames = []
    if triggered:
        stop_time = datetime.timedelta(
            seconds=(frame.timestamp + frame.duration))
        vad_stop.append(str(stop_time))

    if voiced_frames:
        yield b''.join([f.bytes for f in voiced_frames])

    write_csv(file_name, side, vad_start, vad_stop, token_fileID)


def worker(item):

    vad_mode, audio_file, token_fileID = item
    secret = token_fileID.split(".")[0]
    file_name = os.path.basename(audio_file)

    basename = ".".join(file_name.split(".")[:-1])

    fileName = audio_file.split("/")[-2]
    audio, sample_rate = read_wave(audio_file)
    vad = webrtcvad.Vad(int(vad_mode))
    frames = frame_generator(30, audio, sample_rate)
    frames = list(frames)
    segments = vad_collector(
        fileName, basename, sample_rate, 30, 300, vad, frames, token_fileID)

    for i, segment in enumerate(segments):
        path = 'chunk-%002d.wav' % (i,)


def main(args: argparse.Namespace, token_fileID):
    # Create the output directory, if it does not exist.
    print("Creating the output directory...")
    secret = token_fileID.split(".")[0]
    if not os.path.exists(path_to_output+"/"+secret):
        os.makedirs(path_to_output+"/"+secret)
    else:
        print("VAD output directory already exists.")

    # Recursively get list of all audio filenames and their desired output filename.
    workload = []
    print("Gathering input filenames...")

    for path, _, folder in os.walk(path_to_tran+"/"+secret):
        # fileName = path.split("/")[-1]
        fileName = os.path.basename(path)
        if ".DS_Store" in folder:
            folder.remove(".DS_Store")
        global temp
        temp = 0
        for file in folder:
            # Get the base filename without the audio extension.
            basename = ".".join(file.split(".")[:-1])

            audio_file = os.path.join(
                f"{path_to_tran}/{secret}/", f"{fileName}/{file}")
            exist_file = os.path.join(
                f"{path_to_output}/{secret}/", f"{fileName}/{basename}.csv")

        # If we're continuing an interrupted job, we should skip any completed files.
            if args.resume:
                if os.path.exists(exist_file):
                    continue

            workload.append([2, audio_file, token_fileID])
    print(f"Found {len(workload)} files.")

    # Create multiple threads.
    print(f"Starting {args.n_threads} threads...")
    with multiprocessing.Pool(args.n_threads) as pool:
        # r = list(tqdm.tqdm(p.imap(_foo, range(30)), total=30))
        r = list(tqdm(pool.imap(worker, workload)))
        # pool.map(worker, workload)


def splitFile(fileNameOriginal, secret):
    fileName = ""
    for alphabet in fileNameOriginal:
        if alphabet == " ":
            fileName += ""
        else:
            fileName += alphabet
    name = ".".join(fileName.split(".")[:-1])
    signature = fileName.split(".")[-1]

    trandir = path_to_tran + "/"+token + "/" + name

    if not os.path.exists(path_to_tran + "/"+token):
        os.mkdir(path_to_tran + "/"+token)
    if not os.path.exists(trandir):
        os.mkdir(trandir)

    outdir = path_to_output+"/"+token+"/" + name
    if not os.path.exists(path_to_output+"/"+token):
        os.mkdir(path_to_output+"/"+token)
    if not os.path.exists(outdir):
        os.mkdir(outdir)

    p = subprocess.call(
        f"ffmpeg -i {path_to_input}/{token}/{name}.{signature} -map_channel 0.0.0 {trandir}/psychotherapist.wav -map_channel 0.0.1 {trandir}/patient.wav", shell=True)


# parser = argparse.ArgumentParser()
# parser.add_argument("-t", "--token", required=True, help="token of the user")
# arg = vars(parser.parse_args())
# token = arg["token"]
if __name__ == '__main__':

    parser = argparse.ArgumentParser()

    parser.add_argument(
        "--n_threads",
        default=2,
        type=int,
        help="Number of simultaneous VAD calls. Note, all available threads are always used.",
    )
    parser.add_argument(
        "--resume",
        action="store_true",
        help="If True, skip files already completed. Useful for resuming an interrupted job.",
    )
    parser.add_argument("-t", "--token", required=True,
                        help="token of the user")
    parser.add_argument("-i", "--fileid", required=True,
                        help="fileID of the user")

    arg = vars(parser.parse_args())
    token = arg["token"]
    fileID = arg["fileid"]
    # mix = token + "." + fileID

    token_fileID = token + "." + fileID

    args = parser.parse_args()

    for path, _, filenames in os.walk(path_to_input+"/"+token):
        if ".DS_Store" in filenames:
            filenames.remove(".DS_Store")
        for i in filenames:
            temp = ".".join(i.split(".")[:-1])
            if temp[-9:] == fileID:
                splitFile(i, token)
                # os.remove(path_to_input+"/"+token+"/"+i)
        if ".DS_Store" in filenames:
            filenames.remove(".DS_Store")
        if len(filenames) == 0:
            os.rmdir(path_to_input+"/"+token)

    main(args, token_fileID)
    print("Done...")
    write_result_csv(token_fileID)
