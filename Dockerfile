FROM miktex/miktex
RUN apt-get update 
RUN apt-get install wget git language-pack-gnome-zh* vim  -y && \
	cd /root && \
	wget -c https://github.com/jgm/pandoc/releases/download/2.7.2/pandoc-2.7.2-linux.tar.gz && \
	tar -xf pandoc-2.7.2-linux.tar.gz && \
	cp pandoc-2.7.2/bin/pandoc /usr/bin && \
	chmod +x /usr/bin/pandoc
RUN cd /root && \ 
	git clone https://github.com/happy2048/genpdf2.git && \
	cp genpdf2/template.tex /root/template.tex && \
	cp genpdf2/genpdf-server /usr/local/bin && \
	chmod +x /usr/local/bin/genpdf-server && \
	cp -ar genpdf2/fonts /usr/share && \
	
#ADD template.tex /root/template.tex
#ADD genpdf-server /usr/bin/genpdf-server
#ADD fonts /usr/share/fonts
RUN chmod +x /usr/bin/genpdf-server
ENV GOROOT /usr/local/go
ENV PATH /usr/local/go/bin:$PATH
CMD ["/usr/bin/genpdf-server"]

