#!/bin/bash 
CC="msp430-gcc -mmcu=msp430f1611" cmake -DCMAKE_SYSTEM_NAME=Generic -DARITH=msp-asm -DALIGN=2 -DARCH=MSP -DBENCH=1 "-DBN_METHD=BASIC;MULTP;MONTY;BASIC;BASIC;BASIC" -DCHECK=OFF -DCOLOR=OFF "-DCOMP:STRING=-O2 -g -mmcu=msp430f1611 -ffunction-sections -fdata-sections -fno-inline -mdisable-watchdog" -DDOCUM=OFF -DEP_PRECO=OFF "-DFP_METHD=FP_METHD:STRING=BASIC;COMBA;MULTP;MONTY;MONTY;SLIDE" "-DLINK=-Wl,--gc-sections" "-DPP_METHD=INTEG;INTEG;BASIC;OATEP" -DSEED=ZERO -DSHLIB=OFF -DSTRIP=ON -DTESTS=1 -DTIMER=CYCLE -DVERBS=OFF -DWORD=16 -DFP_PRIME=254 -DFP_QNRES=ON -DBN_PRECI=256 -DMD_METHD=SH256 "-DWITH=FP;EP;EPX;PP;PC;DV;CP;MD;BN" -DEC_METHD=PRIME -DPC_METHD=PRIME -DRAND=FIPS $1
