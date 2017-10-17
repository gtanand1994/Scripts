{\rtf1\ansi\ansicpg1252\cocoartf1504\cocoasubrtf830
{\fonttbl\f0\fnil\fcharset0 Monaco;}
{\colortbl;\red255\green255\blue255;}
{\*\expandedcolortbl;;}
\paperw11900\paperh16840\margl1440\margr1440\vieww10800\viewh8400\viewkind0
\pard\tx560\tx1120\tx1680\tx2240\tx2800\tx3360\tx3920\tx4480\tx5040\tx5600\tx6160\tx6720\pardirnatural\partightenfactor0

\f0\fs22 \cf0 \CocoaLigature0 -0-root@nk11p00im-adminm001:~ # cat nested_grep.sh\
#!/bin/bash\
\
BOLD='\\033[1m'\
UNDERLINE='\\033[4m'\
PURPLE='\\033[01;35m'\
GREEN='\\033[01;32m'\
\
file_name=$1;\
string=$2;\
i=1;\
line1=$file_name;\
line2="";\
\
echo -e  "$\{BOLD\}Searching for $string...";\
tput sgr0;\
function search_for_string()\
\{\
\
if [ ! -z "$line1" ];\
then\
\
echo -e "$\{PURPLE\}========================== Search Branch : $i ==========================";\
tput sgr0;\
echo -e "$\{UNDERLINE\}Searching below files for string $string :\\n" \
tput sgr0;\
echo $line1 | sed 's/\\s\\+/\\n/g';\
echo -e "\\n";\
echo "------------------------------------------------------------------------";\
for line in `echo $line1`;do\
if [ -f "$line" ];then\
awk ' \{ print FILENAME" : "FNR" : "$0 \} ' $line | grep -i "$string" 2> /dev/null && echo "------------------------------------------------------------------------";\
fi\
done\
i=$((i+1));\
search_for_files;\
\
else\
echo -e "$\{GREEN\}==========================Search Complete==========================";\
tput sgr0;\
fi\
\}\
\
function search_for_files()\
\{\
line3=`echo $line1 | tr -s ' '| sed 's/ /\\|/g'`;\
line2=`cat $line1 | egrep -v "$line1" | egrep -o '/.*.sh |/.*.pl' |sort -u |sed ':a;N;$!ba;s/\\n/ /g'`;\
line1=$line2;\
search_for_string;\
\}\
\
search_for_string;\
}