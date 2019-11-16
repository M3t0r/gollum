from itertools import chain

from sphinx.domains import Domain
import sphinx.addnodes as sphinx_nodes
from sphinx.util.docutils import SphinxDirective
from docutils.statemachine import ViewList
import docutils.nodes


ENV_KEY = 'gollum_hierarchy_stack'


class GollumDomain(Domain):
    label = "Gollum"
    name = "gollum"

class Plugin(SphinxDirective):
    required_arguments = 1  # the plugin name
    final_argument_whitespace = True  # The option name might contain spaces
    has_content = True

    @property
    def active_plugin(self):
        return self.env.temp_data.setdefault(ENV_KEY, DottedStack())

    def run(self):
        nodes = []

        name = self.arguments[0]
        plugin_type, _, simple_name = name.partition('.')
        full_name = self.active_plugin + name
        index_name = f"{plugin_type}; {simple_name}"

        signature = sphinx_nodes.desc_signature(
            '', '',
            sphinx_nodes.desc_name('', name),
            names=[name, full_name],
            fullname=full_name,
            ids=[full_name],
            module=self.active_plugin,
            first=False,
            noindex=False
        )
        nodes.append(signature)

        # parse the content
        self.active_plugin.append(name)
        description = sphinx_nodes.desc_content()
        self.state.nested_parse(self.content, 0, description)
        nodes.append(description)
        self.active_plugin.pop(name)

        description = sphinx_nodes.desc('', *nodes, objtype="function", domain=GollumDomain.name, desctype="function")
        index = sphinx_nodes.index(entries=[('pair', index_name, full_name, '', None)])
        return [index, description]


class PluginOption(SphinxDirective):
    required_arguments = 1  # the Option name
    final_argument_whitespace = True  # The option name might contain spaces
    has_content = True
    option_spec = {
        "default": str,
        "unit": str,
        "from": str
    }

    @property
    def active_plugin(self):
        return self.env.temp_data.setdefault(ENV_KEY, DottedStack())

    def run(self):
        nodes = []

        name = self.arguments[0]
        full_name = self.active_plugin + name
        index_name = name
        if len(self.active_plugin) > 0:
            index_name = f"{self.active_plugin}; {name}"

        additional_infos = []
        if 'default' in self.options:
            additional_infos.append(f"default: {self.options['default']}")
        if 'unit' in self.options:
            additional_infos.append(f"unit: {self.options['unit']}")
        if 'from' in self.options:
            additional_infos.append(f"from: {self.options['from']}")
        additional_info = ' (' + ', '.join(additional_infos) + ')'
        if len(additional_infos) == 0:
            additional_info = ""

        signature = sphinx_nodes.desc_signature(
            '', '',
            sphinx_nodes.desc_name('', name),
            sphinx_nodes.desc_annotation('', additional_info),
            names=[name, full_name],
            fullname=full_name,
            ids=[full_name],
            first=True
        )
        nodes.append(signature)

        # parse the content
        description = sphinx_nodes.desc_content()
        self.state.nested_parse(self.content, 0, description)
        nodes.append(description)

        description = sphinx_nodes.desc(
            '',
            *nodes,
            objtype="function",
            domain=GollumDomain.name,
            desctype="function"
        )

        index = sphinx_nodes.index(entries=[('single', index_name, full_name, '', None)])
        return [description]


class DottedStack(list):
    def __init__(self, *chunks):
        for chunk in chunks:
            self.append(chunk)

    def append(self, segments):
        for segment in filter(lambda s: len(s) != 0, segments.split('.')):
            super().append(segment)

    def pop(self, segment_string = None):
        if segment_string is None:
            return super().pop()

        segments = DottedStack(segment_string)
        while len(segments) > 0:
            # pop one and compare
            s = segments.pop()
            assert self[-1] == s
            self.pop()

    def __add__(self, segment):
        if len(self) == 0:
            return segment
        if len(segment) == 0:
            return str(self)
        return str(self) + '.' + str(segment)

    def __str__(self):
        return '.'.join(self)

def test_dotted_stack():
    ds = DottedStack("a.b", ".c.d")
    assert str(ds) == 'a.b.c.d'

    last = ds.pop()
    assert last == 'd'
    assert str(ds) == 'a.b.c'

    ds.append('s1..s3.')
    assert str(ds) == 'a.b.c.s1.s3'

    assert 's1' in ds
    assert ds + "blubber" == 'a.b.c.s1.s3.blubber'

    ds.pop("s1..s3")
    assert str(ds) == 'a.b.c'

def setup(app):
    test_dotted_stack()
    app.add_domain(GollumDomain)
    app.add_directive_to_domain("gollum", "plugin", Plugin)
    app.add_directive_to_domain("gollum", "option", PluginOption)
