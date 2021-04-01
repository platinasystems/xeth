dots := $(wildcard *.dot)
svgs := $(patsubst %.dot,%.svg,$(dots))

default: $(svgs)

$(svgs): %.svg: %.dot ; dot -Tsvg -o $@ $<

clean: ; @rm -f $(svgs)
