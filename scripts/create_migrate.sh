#!/bin/sh

# $1 means: migration name

migrate create -ext sql -dir migrations -seq  $1