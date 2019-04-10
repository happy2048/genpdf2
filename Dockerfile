FROM miktex/miktex
RUN apt-get update 
RUN apt-get install wget git language-pack-gnome-zh* vim  -y && \
	cd /root && \
	wget -c https://github.com/jgm/pandoc/releases/download/2.7.2/pandoc-2.7.2-linux.tar.gz && \
	tar -xf pandoc-2.7.2-linux.tar.gz && \
	cp pandoc-2.7.2/bin/pandoc /usr/bin && \
	chmod +x /usr/bin/pandoc
ADD template.tex /root/template.tex
ADD genpdf-server /usr/bin/genpdf-server
ADD fonts /usr/share/fonts
RUN chmod +x /usr/bin/genpdf-server
ENV GOROOT /usr/local/go
ENV PATH /usr/local/go/bin:$PATH
CMD ["/usr/bin/genpdf-server"]

