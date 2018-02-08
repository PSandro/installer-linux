#!/bin/bash
./startup && \
screen -dmS sinusbot ./opt/sinusbot/sinusbot && \
./test
