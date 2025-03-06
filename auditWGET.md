## Audit Questions for WGET

Try to run the following command <code>./go-wget https://pbs.twimg.com/media/EMtmPFLWkAA8CIS.jpg</code>
Did the program download the file EMtmPFLWkAA8CIS.jpg?
Try to run the following command with a link at your choice <code>./go-wget <https://link_of_your_choice.com></code>
Did the program download the expected file?
Try to run the following command <code>./go-wget https://golang.org/dl/go1.16.3.linux-amd64.tar.gz</code>
Did the program download the file go1.16.3.linux-amd64.tar.gz?
Did the program displayed the start time?
Did the start time and the end time respected the format? (yyyy-mm-dd hh:mm:ss)
Did the program displayed the status of the response? (200 OK)
Did the Program displayed the content length of the download?
Is the content length displayed as raw (bytes) and rounded (Mb or Gb)?
Did the program displayed the name and path of the file that was saved?
Try to download a big file, for example: <code>./go-wget https://assets.01-edu.org/wgetDataSamples/Sample.zip</code>
Did the program download the expected file?
While downloading, did the progress bar show the amount that is being downloaded? (KiB or MiB)
While downloading, did the progress bar show the percentage that is being downloaded?
While downloading, did the progress bar show the time that remains to finish the download?
While downloading, did the progress bar progressed smoothly (kept up with the time that the download took to finish)?
Try to run the following command, <code>./go-wget -O=test_20MB.zip https://assets.01-edu.org/wgetDataSamples/20MB.zip</code>
Did the program downloaded the file with the name test_20MB.zip?
Try to run the following command, <code>./go-wget -O=test_20MB.zip -P=~/Downloads/ https://assets.01-edu.org/wgetDataSamples/20MB.zip</code>
Can you see the expected file in the "~/Downloads/" folder?
Try to run the following command, <code>./go-wget --rate-limit=300k https://assets.01-edu.org/wgetDataSamples/20MB.zip</code>
Was the download speed always lower than 300KB/s?
Try to run the following command, <code>./go-wget --rate-limit=700k https://assets.01-edu.org/wgetDataSamples/20MB.zip</code>
Was the download speed always lower than 700KB/s?
Try to run the following command, <code>./go-wget --rate-limit=2M https://assets.01-edu.org/wgetDataSamples/20MB.zip</code>
Was the download speed always lower than 2MB/s?
Try to create a text file with the name downloads.txt and save into it the links below. Then run the command <code>./go-wget -i=downloads.txt</code>

https://assets.01-edu.org/wgetDataSamples/Image_20MB.zip
https://assets.01-edu.org/wgetDataSamples/20MB.zip
https://assets.01-edu.org/wgetDataSamples/Image_10MB.zip

Did the program download all the files from the downloads.txt file? (EMtmPFLWkAA8CIS.jpg, 20MB.zip, 10MB.zip)
Did the downloads occurred in an asynchronous way? (tip: look to the download order)
Try to run the following command, <code>./go-wget -B https://assets.01-edu.org/wgetDataSamples/20MB.zip</code>

Output will be written to ‘wget-log’.

Did the program output the statement above?
Was the download made in "silence" (without displaying anything to the terminal)?
Try and open the log file wget-log.

```
start at <date that the download started>
sending request, awaiting response... status 200 OK
content size: <56370 [~0.06MB]>
saving file to: ./<name-of-the-file-downloaded>
Downloaded [<link-downloaded>]
finished at <date that the download finished>
```

Is the structure of the file organized like above?
And if so, was the file actually downloaded?

## Mirror / Clone a website

Try to run the following command <code>./go-wget --mirror --convert-links http://corndog.io/</code> then try to open the index.html with a browser
Is the site working?
Try to run the following command <code>./go-wget --mirror https://oct82.com/</code> then try to open the index.html with a browser
Is the site working?
Try to run the following command <code>./go-wget --mirror --reject=gif https://oct82.com/</code> then try to open the index.html with a browser
Did the program download the site without the GIFs?
Try to run the following command <code>./go-wget --mirror https://trypap.com/</code> then use the command ls to see the file system of the created folder.

css img index.html

Does the created folder has the same fs as above?
Try to run the following command <code>./go-wget --mirror -X=/img https://trypap.com/</code> then use the command ls to see the file system of the created folder.

css index.html

Does the created folder has the files above?
Try to run the following command <code>./go-wget --mirror https://theuselessweb.com/</code>
Is the site working?
Try to run the following command to mirror a website at your choice <code>./go-wget --mirror <https://link_of_your_choice.com></code>
Did the program mirror the website?
