#!/bin/sh

migrate -path migrations -database "postgres://socialqueue:socialqueue@localhost:5432/?sslmode=disable" up